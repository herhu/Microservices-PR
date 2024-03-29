package grpc

import (
	"context"
	"time"

	"github.com/go-playground/validator"
	"github.com/herhu/Microservices-PR/pkg/logger"
	"github.com/herhu/Microservices-PR/pkg/tracing"
	"github.com/herhu/Microservices-PR/pkg/utils"
	"github.com/herhu/Microservices-PR/reader_service/config"
	"github.com/herhu/Microservices-PR/reader_service/internal/metrics"
	"github.com/herhu/Microservices-PR/reader_service/internal/models"
	"github.com/herhu/Microservices-PR/reader_service/internal/product/commands"
	"github.com/herhu/Microservices-PR/reader_service/internal/product/queries"
	"github.com/herhu/Microservices-PR/reader_service/internal/product/service"
	readerService "github.com/herhu/Microservices-PR/reader_service/proto/product_reader"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcService struct {
	log     logger.Logger
	cfg     *config.Config
	v       *validator.Validate
	ps      *service.ProductService
	metrics *metrics.ReaderServiceMetrics
}

func NewReaderGrpcService(log logger.Logger, cfg *config.Config, v *validator.Validate, ps *service.ProductService, metrics *metrics.ReaderServiceMetrics) *grpcService {
	return &grpcService{log: log, cfg: cfg, v: v, ps: ps, metrics: metrics}
}

func (s *grpcService) CreateProduct(ctx context.Context, req *readerService.CreateProductReq) (*readerService.CreateProductRes, error) {
	s.metrics.CreateProductGrpcRequests.Inc()

	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.CreateProduct")
	defer span.Finish()

	command := commands.NewCreateProductCommand(req.GetProductID(), req.GetName(), req.GetDescription(), req.GetPrice(), time.Now(), time.Now())
	if err := s.v.StructCtx(ctx, command); err != nil {
		s.log.WarnMsg("validate", err)
		return nil, s.errResponse(codes.InvalidArgument, err)
	}

	if err := s.ps.Commands.CreateProduct.Handle(ctx, command); err != nil {
		s.log.WarnMsg("CreateProduct.Handle", err)
		return nil, s.errResponse(codes.InvalidArgument, err)
	}

	s.metrics.SuccessGrpcRequests.Inc()
	return &readerService.CreateProductRes{ProductID: req.GetProductID()}, nil
}

func (s *grpcService) UpdateProduct(ctx context.Context, req *readerService.UpdateProductReq) (*readerService.UpdateProductRes, error) {
	s.metrics.UpdateProductGrpcRequests.Inc()

	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.UpdateProduct")
	defer span.Finish()

	command := commands.NewUpdateProductCommand(req.GetProductID(), req.GetName(), req.GetDescription(), req.GetPrice(), time.Now())
	if err := s.v.StructCtx(ctx, command); err != nil {
		s.log.WarnMsg("validate", err)
		return nil, s.errResponse(codes.InvalidArgument, err)
	}

	if err := s.ps.Commands.UpdateProduct.Handle(ctx, command); err != nil {
		s.log.WarnMsg("UpdateProduct.Handle", err)
		return nil, s.errResponse(codes.InvalidArgument, err)
	}

	s.metrics.SuccessGrpcRequests.Inc()
	return &readerService.UpdateProductRes{ProductID: req.GetProductID()}, nil
}

func (s *grpcService) GetProductById(ctx context.Context, req *readerService.GetProductByIdReq) (*readerService.GetProductByIdRes, error) {
	s.metrics.GetProductByIdGrpcRequests.Inc()

	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.GetProductById")
	defer span.Finish()

	productUUID, err := uuid.FromString(req.GetProductID())
	if err != nil {
		s.log.WarnMsg("uuid.FromString", err)
		return nil, s.errResponse(codes.InvalidArgument, err)
	}

	query := queries.NewGetProductByIdQuery(productUUID)
	if err := s.v.StructCtx(ctx, query); err != nil {
		s.log.WarnMsg("validate", err)
		return nil, s.errResponse(codes.InvalidArgument, err)
	}

	product, err := s.ps.Queries.GetProductById.Handle(ctx, query)
	if err != nil {
		s.log.WarnMsg("GetProductById.Handle", err)
		return nil, s.errResponse(codes.Internal, err)
	}

	s.metrics.SuccessGrpcRequests.Inc()
	return &readerService.GetProductByIdRes{Product: models.ProductToGrpcMessage(product)}, nil
}

func (s *grpcService) SearchProduct(ctx context.Context, req *readerService.SearchReq) (*readerService.SearchRes, error) {
	s.metrics.SearchProductGrpcRequests.Inc()

	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.SearchProduct")
	defer span.Finish()

	pq := utils.NewPaginationQuery(int(req.GetSize()), int(req.GetPage()))

	query := queries.NewSearchProductQuery(req.GetSearch(), pq)
	productsList, err := s.ps.Queries.SearchProduct.Handle(ctx, query)
	if err != nil {
		s.log.WarnMsg("SearchProduct.Handle", err)
		return nil, s.errResponse(codes.Internal, err)
	}

	s.metrics.SuccessGrpcRequests.Inc()
	return models.ProductListToGrpc(productsList), nil
}

func (s *grpcService) DeleteProductByID(ctx context.Context, req *readerService.DeleteProductByIdReq) (*readerService.DeleteProductByIdRes, error) {
	s.metrics.DeleteProductGrpcRequests.Inc()

	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.DeleteProductByID")
	defer span.Finish()

	productUUID, err := uuid.FromString(req.GetProductID())
	if err != nil {
		s.log.WarnMsg("uuid.FromString", err)
		return nil, s.errResponse(codes.InvalidArgument, err)
	}

	if err := s.ps.Commands.DeleteProduct.Handle(ctx, commands.NewDeleteProductCommand(productUUID)); err != nil {
		s.log.WarnMsg("DeleteProduct.Handle", err)
		return nil, s.errResponse(codes.Internal, err)
	}

	s.metrics.SuccessGrpcRequests.Inc()
	return &readerService.DeleteProductByIdRes{}, nil
}

func (s *grpcService) errResponse(c codes.Code, err error) error {
	s.metrics.ErrorGrpcRequests.Inc()
	return status.Error(c, err.Error())
}
