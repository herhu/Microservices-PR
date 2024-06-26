package commands

import (
	"context"

	"github.com/herhu/Microservices-PR/pkg/logger"
	"github.com/herhu/Microservices-PR/reader_service/config"
	"github.com/herhu/Microservices-PR/reader_service/internal/models"
	"github.com/herhu/Microservices-PR/reader_service/internal/product/repository"
	"github.com/opentracing/opentracing-go"
)

type CreateProductCmdHandler interface {
	Handle(ctx context.Context, command *CreateProductCommand) error
}

type createProductHandler struct {
	log       logger.Logger
	cfg       *config.Config
	mongoRepo repository.Repository
	redisRepo repository.CacheRepository
}

func NewCreateProductHandler(log logger.Logger, cfg *config.Config, mongoRepo repository.Repository, redisRepo repository.CacheRepository) *createProductHandler {
	return &createProductHandler{log: log, cfg: cfg, mongoRepo: mongoRepo, redisRepo: redisRepo}
}

func (c *createProductHandler) Handle(ctx context.Context, command *CreateProductCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "createProductHandler.Handle")
	defer span.Finish()

	product := &models.Product{
		ProductID:   command.ProductID,
		Name:        command.Name,
		Description: command.Description,
		Price:       command.Price,
		CreatedAt:   command.CreatedAt,
		UpdatedAt:   command.UpdatedAt,
	}

	created, err := c.mongoRepo.CreateProduct(ctx, product)
	if err != nil {
		return err
	}

	c.redisRepo.PutProduct(ctx, created.ProductID, created)
	return nil
}
