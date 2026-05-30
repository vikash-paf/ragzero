package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"github.com/vikash-paf/ragzero/internal/handler"
	"github.com/vikash-paf/ragzero/internal/middleware"
	"github.com/vikash-paf/ragzero/internal/models"
	"github.com/vikash-paf/ragzero/internal/repository"
	"github.com/vikash-paf/ragzero/internal/service"
	"github.com/vikash-paf/ragzero/internal/telemetry"
)

func main() {
	ctx := context.Background()

	// Initialize Structured Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Initialize Telemetry
	tp, err := telemetry.InitTracer(ctx, "ragzero-api", os.Stderr)
	if err != nil {
		log.Fatal("failed to initialize tracer: ", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Load API Keys
	if err := middleware.LoadAPIKeys("config/api_keys.json"); err != nil {
		slog.Error("Failed to load API keys", "error", err)
		os.Exit(1)
	}

	// 1. Database Connection (PostgreSQL)
	pgDSN := os.Getenv("DATABASE_URL")
	if pgDSN == "" {
		pgDSN = "postgres://user:password@localhost:5432/ragzero?sslmode=disable"
	}
	pgPool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		slog.Error("Unable to connect to database", "error", err)
		os.Exit(1)
	}
	defer pgPool.Close()

	if err := repository.RunMigrations(pgDSN); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// 2. Cache Connection (Redis)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer rdb.Close()

	// 3. Search Connection (Elasticsearch)
	esAddr := os.Getenv("ELASTICSEARCH_URL")
	if esAddr == "" {
		esAddr = "http://localhost:9200"
	}
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{esAddr},
	})
	if err != nil {
		slog.Error("Unable to connect to Elasticsearch", "error", err)
		os.Exit(1)
	}

	esIndex := "documents"
	if err := repository.InitIndex(ctx, esClient, esIndex); err != nil {
		slog.Warn("Failed to initialize Elasticsearch index", "error", err)
	}

	// 4. Initialize Repositories
	docRepo := repository.NewPostgresDocumentRepository(pgPool)
	cacheRepo := repository.NewRedisCacheRepository(rdb)
	searchRepo := repository.NewElasticsearchRepository(esClient, esIndex)

	// 5. Initialize Services & Workers
	indexChan := make(chan *models.Document, 100)
	indexer := service.NewIndexerWorker(searchRepo, indexChan)
	go indexer.Start(ctx)

	docSvc := service.NewDocumentService(docRepo, searchRepo, cacheRepo, indexChan)

	// 6. Setup API
	r := gin.New() // Use New() to avoid default middleware if we want full control
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("ragzero-api"))

	h := handler.NewDocumentHandler(docSvc)

	// Health check
	r.GET("/health", h.HandleHealthCheck)

	// API Routes with Middleware
	api := r.Group("/")
	api.Use(middleware.AuthMiddleware())
	api.Use(middleware.RateLimitMiddleware(cacheRepo, 100))
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

	slog.Info("Starting server", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("Failed to run server", "error", err)
		os.Exit(1)
	}
}
