package service

import (
	"context"
	"log"

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
	log.Println("Indexer worker started...")
	for {
		select {
		case doc, ok := <-w.indexChan:
			if !ok {
				return
			}
			if err := w.searchRepo.Index(ctx, doc); err != nil {
				log.Printf("Failed to index document %s: %v", doc.ID, err)
			}
		case <-ctx.Done():
			return
		}
	}
}
