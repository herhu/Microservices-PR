package kafka

import (
	"context"
	"sync"

	"github.com/go-playground/validator"
	"github.com/herhu/Microservices-PR/pkg/logger"
	"github.com/herhu/Microservices-PR/writer_service/config"
	"github.com/herhu/Microservices-PR/writer_service/internal/metrics"
	"github.com/herhu/Microservices-PR/writer_service/internal/product/service"
	"github.com/segmentio/kafka-go"
)

const (
	PoolSize = 30
)

type productMessageProcessor struct {
	log     logger.Logger
	cfg     *config.Config
	v       *validator.Validate
	ps      *service.ProductService
	metrics *metrics.WriterServiceMetrics
}

func NewProductMessageProcessor(log logger.Logger, cfg *config.Config, v *validator.Validate, ps *service.ProductService, metrics *metrics.WriterServiceMetrics) *productMessageProcessor {
	return &productMessageProcessor{log: log, cfg: cfg, v: v, ps: ps, metrics: metrics}
}

func (s *productMessageProcessor) ProcessMessages(ctx context.Context, r *kafka.Reader, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		m, err := r.FetchMessage(ctx)
		if err != nil {
			s.log.Warnf("workerID: %v, err: %v", workerID, err)
			continue
		}

		s.logProcessMessage(m, workerID)

		switch m.Topic {
		case s.cfg.KafkaTopics.ProductCreate.TopicName:
			s.processCreateProduct(ctx, r, m)
		case s.cfg.KafkaTopics.ProductUpdate.TopicName:
			s.processUpdateProduct(ctx, r, m)
		case s.cfg.KafkaTopics.ProductDelete.TopicName:
			s.processDeleteProduct(ctx, r, m)
		}
	}
}
