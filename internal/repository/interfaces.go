package repository

import (
	"context"
	"github.com/vikash-paf/ragzero/internal/models"
)

type DocumentRepository interface {
	Save(ctx context.Context, doc *models.Document) error
	GetByID(ctx context.Context, id string, tenantID string) (*models.Document, error)
	Delete(ctx context.Context, id string, tenantID string) error
}

type SearchRepository interface {
	Index(ctx context.Context, doc *models.Document) error
	Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error)
	Delete(ctx context.Context, id string, tenantID string) error
}

type CacheRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl int) error
	Incr(ctx context.Context, key string, ttl int) (int64, error)
}
