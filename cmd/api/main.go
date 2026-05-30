package main

import (
	"context"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/vikash-paf/ragzero/internal/handler"
	"github.com/vikash-paf/ragzero/internal/middleware"
	"github.com/vikash-paf/ragzero/internal/models"
	"github.com/vikash-paf/ragzero/internal/repository"
	"github.com/vikash-paf/ragzero/internal/service"
)

func main() {
	ctx := context.Background()

	pgDSN := os.Getenv("DATABASE_URL")
	if pgDSN == "" {
	}
	pgPool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
	defer pgPool.Close()

	if err := repository.RunMigrations(pgDSN); err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer rdb.Close()

	esAddr := os.Getenv("ELASTICSEARCH_URL")
	if esAddr == "" {
	}
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esAddr},
	})
	if err != nil {
		log.Fatal("Unable to connect to Elasticsearch: ", err)
	}

	esIndex := "documents"
	if err := repository.InitIndex(ctx, esClient, esIndex); err != nil {
		log.Printf("Warning: Failed to initialize Elasticsearch index (might already exist): %v", err)
	}

	docRepo := repository.NewPostgresDocumentRepository(pgPool)
	cacheRepo := repository.NewRedisCacheRepository(rdb)
	searchRepo := repository.NewElasticsearchRepository(esClient, esIndex)

	indexChan := make(chan *models.Document, 100)
	indexer := service.NewIndexerWorker(searchRepo, indexChan)
	go indexer.Start(ctx)

	docSvc := service.NewDocumentService(docRepo, searchRepo, cacheRepo, indexChan)

	r := gin.Default()
	h := handler.NewDocumentHandler(docSvc)

	r.GET("/health", h.HandleHealthCheck)

	api := r.Group("/")
	api.Use(middleware.TenantMiddleware())
	{
		api.POST("/documents", h.HandleCreateDocument)
		api.GET("/search", h.HandleSearchDocuments)
		api.GET("/documents/:id", h.HandleGetDocument)
		api.DELETE("/documents/:id", h.HandleDeleteDocument)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
