package models

import "time"

type Document struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchRequest struct {
	Query    string `json:"query"`
	TenantID string `json:"tenant_id"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

type SearchResult struct {
	ID        string              `json:"id"`
	Title     string              `json:"title"`
	Score     float64             `json:"score"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int64          `json:"total"`
	TookMS  int64          `json:"took_ms"`
}
