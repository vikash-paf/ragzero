package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vikash-paf/ragzero/internal/models"
)

// Mocks
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Save(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetByID(ctx context.Context, id string, tenantID string) (*models.Document, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentRepository) Delete(ctx context.Context, id string, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

type MockSearchRepository struct {
	mock.Mock
}

func (m *MockSearchRepository) Index(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockSearchRepository) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SearchResponse), args.Error(1)
}

func (m *MockSearchRepository) Delete(ctx context.Context, id string, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value string, ttl int) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) Incr(ctx context.Context, key string, ttl int) (int64, error) {
	args := m.Called(ctx, key, ttl)
	return args.Get(0).(int64), args.Error(1)
}

func TestDocumentService_CreateDocument(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	searchRepo := new(MockSearchRepository)
	cacheRepo := new(MockCacheRepository)
	indexChan := make(chan *models.Document, 1)

	svc := NewDocumentService(docRepo, searchRepo, cacheRepo, indexChan)

	doc := &models.Document{
		ID:       "test-1",
		TenantID: "tenant-a",
		Title:    "Test Title",
		Content:  "Test Content",
	}

	docRepo.On("Save", mock.Anything, doc).Return(nil)

	err := svc.CreateDocument(context.Background(), doc)

	assert.NoError(t, err)
	docRepo.AssertExpectations(t)
	
	// Check if document was sent to indexChan
	sentDoc := <-indexChan
	assert.Equal(t, doc.ID, sentDoc.ID)
}

func TestDocumentService_SearchDocuments_CacheHit(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	searchRepo := new(MockSearchRepository)
	cacheRepo := new(MockCacheRepository)
	indexChan := make(chan *models.Document, 1)

	svc := NewDocumentService(docRepo, searchRepo, cacheRepo, indexChan)

	req := &models.SearchRequest{
		Query:    "test",
		TenantID: "tenant-a",
	}

	cacheRepo.On("Get", mock.Anything, "search:tenant-a:test").Return(`{"results":[{"id":"test-1","title":"Test"}]}`, nil)

	resp, err := svc.SearchDocuments(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-1", resp.Results[0].ID)
	cacheRepo.AssertExpectations(t)
	searchRepo.AssertNotCalled(t, "Search", mock.Anything, mock.Anything)
}
