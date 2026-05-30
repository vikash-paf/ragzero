package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vikash-paf/ragzero/internal/models"
)

type postgresDocumentRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresDocumentRepository(pool *pgxpool.Pool) DocumentRepository {
	return &postgresDocumentRepo{pool: pool}
}

func (r *postgresDocumentRepo) Save(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (id, tenant_id, title, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			created_at = EXCLUDED.created_at
	`
	_, err := r.pool.Exec(ctx, query, doc.ID, doc.TenantID, doc.Title, doc.Content, doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save document to postgres: %w", err)
	}
	return nil
}

func (r *postgresDocumentRepo) GetByID(ctx context.Context, id string, tenantID string) (*models.Document, error) {
	query := `
		SELECT id, tenant_id, title, content, created_at
		FROM documents
		WHERE id = $1 AND tenant_id = $2
	`
	doc := &models.Document{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&doc.ID, &doc.TenantID, &doc.Title, &doc.Content, &doc.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get document from postgres: %w", err)
	}
	return doc, nil
}

func (r *postgresDocumentRepo) Delete(ctx context.Context, id string, tenantID string) error {
	query := `DELETE FROM documents WHERE id = $1 AND tenant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete document from postgres: %w", err)
	}
	return nil
}

