package service

import (
	kafkaClient "github.com/herhu/Microservices-PR/pkg/kafka"
	"github.com/herhu/Microservices-PR/pkg/logger"
	"github.com/herhu/Microservices-PR/writer_service/config"
	"github.com/herhu/Microservices-PR/writer_service/internal/product/commands"
	"github.com/herhu/Microservices-PR/writer_service/internal/product/queries"
	"github.com/herhu/Microservices-PR/writer_service/internal/product/repository"
)

type ProductService struct {
	Commands *commands.ProductCommands
	Queries  *queries.ProductQueries
}

func NewProductService(log logger.Logger, cfg *config.Config, pgRepo repository.Repository, kafkaProducer kafkaClient.Producer) *ProductService {

	updateProductHandler := commands.NewUpdateProductHandler(log, cfg, pgRepo, kafkaProducer)
	createProductHandler := commands.NewCreateProductHandler(log, cfg, pgRepo, kafkaProducer)
	deleteProductHandler := commands.NewDeleteProductHandler(log, cfg, pgRepo, kafkaProducer)

	getProductByIdHandler := queries.NewGetProductByIdHandler(log, cfg, pgRepo)

	productCommands := commands.NewProductCommands(createProductHandler, updateProductHandler, deleteProductHandler)
	productQueries := queries.NewProductQueries(getProductByIdHandler)

	return &ProductService{Commands: productCommands, Queries: productQueries}
}
