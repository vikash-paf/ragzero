package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"github.com/vikash-paf/ragzero/internal/models"
	"github.com/vikash-paf/ragzero/internal/repository"
)

var tracer = otel.Tracer("document-service")

type DocumentService struct {
	docRepo    repository.DocumentRepository
	searchRepo repository.SearchRepository
	cacheRepo  repository.CacheRepository
	indexChan  chan<- *models.Document
}

func NewDocumentService(
	docRepo repository.DocumentRepository,
	searchRepo repository.SearchRepository,
	cacheRepo repository.CacheRepository,
	indexChan chan<- *models.Document,
) *DocumentService {
	return &DocumentService{
		docRepo:    docRepo,
		searchRepo: searchRepo,
		cacheRepo:  cacheRepo,
		indexChan:  indexChan,
	}
}

func (s *DocumentService) CreateDocument(ctx context.Context, doc *models.Document) error {
	ctx, span := tracer.Start(ctx, "CreateDocument")
	defer span.End()

	doc.CreatedAt = time.Now()
	
	slog.Info("Creating document", "id", doc.ID, "tenant_id", doc.TenantID)

	if err := s.docRepo.Save(ctx, doc); err != nil {
		slog.Error("Failed to save document to primary store", "error", err)
		return err
	}

	select {
	case s.indexChan <- doc:
		slog.Debug("Document sent to async indexer", "id", doc.ID)
	default:
		slog.Warn("Indexer channel full, falling back to goroutine", "id", doc.ID)
		go func(d *models.Document) {
			s.searchRepo.Index(context.Background(), d)
		}(doc)
	}

	return nil
}

func (s *DocumentService) GetDocument(ctx context.Context, id string, tenantID string) (*models.Document, error) {
	ctx, span := tracer.Start(ctx, "GetDocument")
	defer span.End()

	return s.docRepo.GetByID(ctx, id, tenantID)
}

func (s *DocumentService) SearchDocuments(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	ctx, span := tracer.Start(ctx, "SearchDocuments")
	defer span.End()

	slog.Info("Searching documents", "query", req.Query, "tenant_id", req.TenantID)

	// Check L2 Cache
	cacheKey := fmt.Sprintf("search:%s:%s", req.TenantID, req.Query)
	cached, err := s.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var resp models.SearchResponse
		if err := json.Unmarshal([]byte(cached), &resp); err == nil {
			slog.Debug("Search cache hit", "key", cacheKey)
			return &resp, nil
		}
	}

	// Query Elasticsearch
	resp, err := s.searchRepo.Search(ctx, req)
	if err != nil {
		slog.Error("Elasticsearch search failed", "error", err)
		return nil, err
	}

	// Set L2 Cache
	if data, err := json.Marshal(resp); err == nil {
		slog.Debug("Updating search cache", "key", cacheKey)
		s.cacheRepo.Set(ctx, cacheKey, string(data), 300)
	}

	return resp, nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, id string, tenantID string) error {
	ctx, span := tracer.Start(ctx, "DeleteDocument")
	defer span.End()

	slog.Info("Deleting document", "id", id, "tenant_id", tenantID)

	if err := s.docRepo.Delete(ctx, id, tenantID); err != nil {
		return err
	}
	return s.searchRepo.Delete(ctx, id, tenantID)
}
