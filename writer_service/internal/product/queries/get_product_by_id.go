package queries

import (
	"context"

	"github.com/herhu/Microservices-PR/pkg/logger"
	"github.com/herhu/Microservices-PR/writer_service/config"
	"github.com/herhu/Microservices-PR/writer_service/internal/models"
	"github.com/herhu/Microservices-PR/writer_service/internal/product/repository"
)

type GetProductByIdHandler interface {
	Handle(ctx context.Context, query *GetProductByIdQuery) (*models.Product, error)
}

type getProductByIdHandler struct {
	log    logger.Logger
	cfg    *config.Config
	pgRepo repository.Repository
}

func NewGetProductByIdHandler(log logger.Logger, cfg *config.Config, pgRepo repository.Repository) *getProductByIdHandler {
	return &getProductByIdHandler{log: log, cfg: cfg, pgRepo: pgRepo}
}

func (q *getProductByIdHandler) Handle(ctx context.Context, query *GetProductByIdQuery) (*models.Product, error) {
	return q.pgRepo.GetProductById(ctx, query.ProductID)
}
