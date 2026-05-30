# Ragzero: Distributed Document Search Service

Ragzero is a high-performance, multi-tenant document search service built with Go, Elasticsearch, Redis, and PostgreSQL. It is designed to handle millions of documents with sub-500ms latency.

## Architecture Highlights
- **Multi-Tenancy:** Strict isolation using header-based `X-Tenant-Id` enforcement.
- **Eventual Consistency:** Asynchronous indexing via internal worker pools for high write availability.
- **Caching:** L2 query results caching using Redis (Cache-Aside pattern).
- **Rate Limiting:** Per-tenant rate limiting to prevent noisy neighbor issues.
- **Search Engine:** Elasticsearch with BM25 relevance ranking across full document content.

## Prerequisites
- Docker & Docker Compose
- Go 1.26+ (managed via asdf)

## Getting Started

### 1. Start the services
```bash
docker-compose up --build
```
This will start the API, Elasticsearch, Redis, and PostgreSQL.

### 2. Verify Health
```bash
curl http://localhost:8080/health
```

### 3. Usage Examples

#### Index a Document
```bash
curl -X POST http://localhost:8080/documents \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: tenant-a" \
  -d '{
    "id": "doc-1",
    "title": "Distributed Systems",
    "content": "A distributed system is a system whose components are located on different networked computers..."
  }'
```

#### Search Documents
```bash
curl "http://localhost:8080/search?q=distributed" \
  -H "X-Tenant-Id: tenant-a"
```

#### Get Document Details
```bash
curl http://localhost:8080/documents/doc-1 \
  -H "X-Tenant-Id: tenant-a"
```

#### Delete Document
```bash
curl -X DELETE http://localhost:8080/documents/doc-1 \
  -H "X-Tenant-Id: tenant-a"
```

## Production Readiness
For a detailed analysis on scaling to 100x, resilience, and security, please refer to the `architecture.md` file (stored in `.tmp/` for this assessment).
