package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vikash-paf/ragzero/internal/models"
	"github.com/vikash-paf/ragzero/internal/repository"
)

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
	doc.CreatedAt = time.Now()
	
	if err := s.docRepo.Save(ctx, doc); err != nil {
		return err
	}

	select {
	case s.indexChan <- doc:
	default:
		go func(d *models.Document) {
			s.searchRepo.Index(context.Background(), d)
		}(doc)
	}

	return nil
}

func (s *DocumentService) GetDocument(ctx context.Context, id string, tenantID string) (*models.Document, error) {
	return s.docRepo.GetByID(ctx, id, tenantID)
}

func (s *DocumentService) SearchDocuments(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	cacheKey := fmt.Sprintf("search:%s:%s", req.TenantID, req.Query)
	cached, _ := s.cacheRepo.Get(ctx, cacheKey)
	if cached != "" {
		var resp models.SearchResponse
		if err := json.Unmarshal([]byte(cached), &resp); err == nil {
			return &resp, nil
		}
	}

	resp, err := s.searchRepo.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(resp); err == nil {
		s.cacheRepo.Set(ctx, cacheKey, string(data), 300)
	}

	return resp, nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, id string, tenantID string) error {
	if err := s.docRepo.Delete(ctx, id, tenantID); err != nil {
		return err
	}
	return s.searchRepo.Delete(ctx, id, tenantID)
}
