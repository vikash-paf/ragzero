package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/vikash-paf/ragzero/internal/models"
)

type elasticsearchRepo struct {
	client *elasticsearch.Client
	index  string
}

func NewElasticsearchRepository(client *elasticsearch.Client, index string) SearchRepository {
	return &elasticsearchRepo{
		client: client,
		index:  index,
	}
}

func (r *elasticsearchRepo) Index(ctx context.Context, doc *models.Document) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal doc: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      r.index,
		DocumentID: doc.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("es index: %s", res.String())
	}

	return nil
}

func (r *elasticsearchRepo) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":     req.Query,
							"fields":    []string{"title^2", "content"},
							"fuzziness": "AUTO",
						},
					},
				},
				"filter": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"tenant_id": req.TenantID,
						},
					},
				},
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title":   map[string]interface{}{},
				"content": map[string]interface{}{},
			},
		},
		"size": req.Limit,
		"from": req.Offset,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.index),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("es search: %s", res.String())
	}

	var rMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rMap); err != nil {
		return nil, err
	}

	searchRes := &models.SearchResponse{
		Results: []models.SearchResult{},
		TookMS:  int64(rMap["took"].(float64)),
	}

	hits := rMap["hits"].(map[string]interface{})
	searchRes.Total = int64(hits["total"].(map[string]interface{})["value"].(float64))

	for _, hit := range hits["hits"].([]interface{}) {
		item := hit.(map[string]interface{})
		source := item["_source"].(map[string]interface{})
		
		resItem := models.SearchResult{
			ID:    item["_id"].(string),
			Title: source["title"].(string),
			Score: item["_score"].(float64),
		}

		if h, ok := item["highlight"].(map[string]interface{}); ok {
			resItem.Highlight = make(map[string][]string)
			for k, v := range h {
				var highlights []string
				for _, snippet := range v.([]interface{}) {
					highlights = append(highlights, snippet.(string))
				}
				resItem.Highlight[k] = highlights
			}
		}

		searchRes.Results = append(searchRes.Results, resItem)
	}

	return searchRes, nil
}

func (r *elasticsearchRepo) Delete(ctx context.Context, id string, tenantID string) error {
	req := esapi.DeleteRequest{
		Index:      r.index,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("es delete: %s", res.String())
	}

	return nil
}

func InitIndex(ctx context.Context, client *elasticsearch.Client, index string) error {
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		},
		"mappings": {
			"properties": {
				"id": { "type": "keyword" },
				"tenant_id": { "type": "keyword" },
				"title": { "type": "text", "analyzer": "standard" },
				"content": { "type": "text", "analyzer": "standard" },
				"created_at": { "type": "date" }
			}
		}
	}`

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(mapping),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 400 {
		return fmt.Errorf("es create index: %s", res.String())
	}

	return nil
}
