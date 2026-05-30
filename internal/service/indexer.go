package service

import (
	"context"
	"log/slog"

	"github.com/vikash-paf/ragzero/internal/models"
	"github.com/vikash-paf/ragzero/internal/repository"
)

type IndexerWorker struct {
	searchRepo repository.SearchRepository
	indexChan  <-chan *models.Document
}

func NewIndexerWorker(searchRepo repository.SearchRepository, indexChan <-chan *models.Document) *IndexerWorker {
	return &IndexerWorker{
		searchRepo: searchRepo,
		indexChan:  indexChan,
	}
}

func (w *IndexerWorker) Start(ctx context.Context) {
	slog.Info("Indexer worker started")
	for {
		select {
		case doc, ok := <-w.indexChan:
			if !ok {
				return
			}
			if err := w.searchRepo.Index(ctx, doc); err != nil {
				slog.Error("Failed to index document", "id", doc.ID, "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
