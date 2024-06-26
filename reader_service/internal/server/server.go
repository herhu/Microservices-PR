package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator"
	"github.com/go-redis/redis/v8"
	"github.com/herhu/Microservices-PR/pkg/interceptors"
	kafkaClient "github.com/herhu/Microservices-PR/pkg/kafka"
	"github.com/herhu/Microservices-PR/pkg/logger"
	"github.com/herhu/Microservices-PR/pkg/mongodb"
	redisClient "github.com/herhu/Microservices-PR/pkg/redis"
	"github.com/herhu/Microservices-PR/pkg/tracing"
	"github.com/herhu/Microservices-PR/reader_service/config"
	"github.com/herhu/Microservices-PR/reader_service/internal/metrics"
	readerKafka "github.com/herhu/Microservices-PR/reader_service/internal/product/delivery/kafka"
	"github.com/herhu/Microservices-PR/reader_service/internal/product/repository"
	"github.com/herhu/Microservices-PR/reader_service/internal/product/service"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type server struct {
	log         logger.Logger
	cfg         *config.Config
	v           *validator.Validate
	kafkaConn   *kafka.Conn
	im          interceptors.InterceptorManager
	mongoClient *mongo.Client
	redisClient redis.UniversalClient
	ps          *service.ProductService
	metrics     *metrics.ReaderServiceMetrics
}

func NewServer(log logger.Logger, cfg *config.Config) *server {
	return &server{log: log, cfg: cfg, v: validator.New()}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	s.im = interceptors.NewInterceptorManager(s.log)
	s.metrics = metrics.NewReaderServiceMetrics(s.cfg)

	mongoDBConn, err := mongodb.NewMongoDBConn(ctx, s.cfg.Mongo)
	if err != nil {
		return errors.Wrap(err, "NewMongoDBConn")
	}
	s.mongoClient = mongoDBConn
	defer mongoDBConn.Disconnect(ctx) // nolint: errcheck
	s.log.Infof("Mongo connected: %v", mongoDBConn.NumberSessionsInProgress())

	s.redisClient = redisClient.NewUniversalRedisClient(s.cfg.Redis)
	defer s.redisClient.Close() // nolint: errcheck
	s.log.Infof("Redis connected: %+v", s.redisClient.PoolStats())

	mongoRepo := repository.NewMongoRepository(s.log, s.cfg, s.mongoClient)
	redisRepo := repository.NewRedisRepository(s.log, s.cfg, s.redisClient)

	s.ps = service.NewProductService(s.log, s.cfg, mongoRepo, redisRepo)

	readerMessageProcessor := readerKafka.NewReaderMessageProcessor(s.log, s.cfg, s.v, s.ps, s.metrics)

	s.log.Info("Starting Reader Kafka consumers")
	cg := kafkaClient.NewConsumerGroup(s.cfg.Kafka.Brokers, s.cfg.Kafka.GroupID, s.log)
	go cg.ConsumeTopic(ctx, s.getConsumerGroupTopics(), readerKafka.PoolSize, readerMessageProcessor.ProcessMessages)

	if err := s.connectKafkaBrokers(ctx); err != nil {
		return errors.Wrap(err, "s.connectKafkaBrokers")
	}
	defer s.kafkaConn.Close() // nolint: errcheck

	s.runHealthCheck(ctx)
	s.runMetrics(cancel)

	if s.cfg.Jaeger.Enable {
		tracer, closer, err := tracing.NewJaegerTracer(s.cfg.Jaeger)
		if err != nil {
			return err
		}
		defer closer.Close() // nolint: errcheck
		opentracing.SetGlobalTracer(tracer)
	}

	closeGrpcServer, grpcServer, err := s.newReaderGrpcServer()
	if err != nil {
		return errors.Wrap(err, "NewScmGrpcServer")
	}
	defer closeGrpcServer() // nolint: errcheck

	<-ctx.Done()
	grpcServer.GracefulStop()
	return nil
}
