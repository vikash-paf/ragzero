# Ragzero: Distributed Document Search Service

Ragzero is a high-performance, multi-tenant document search service built with Go, Elasticsearch, Redis, and PostgreSQL. It is designed to handle millions of documents with sub-500ms latency.

## Architecture Highlights
- **Multi-Tenancy:** Strict isolation using `X-Tenant-Id` and `X-API-Key` header enforcement.
- **Eventual Consistency:** Asynchronous indexing via internal worker pools for high write availability.
- **Observability:** Distributed tracing with OpenTelemetry and structured JSON logging.
- **Caching:** L2 query results caching using Redis (Cache-Aside pattern).
- **Rate Limiting:** Per-tenant rate limiting to prevent noisy neighbor issues.
- **Search Engine:** Elasticsearch with BM25 relevance ranking, fuzzy matching, and result highlighting.


## Prerequisites
- Docker & Docker Compose
- Go 1.26+ (managed via asdf)

## Getting Started

### 1. Start the services
```bash
make up
```

### 2. API Testing
Sample API requests are available in the [docs/http/](./docs/http/) directory. These can be used to verify system functionality, including fuzzy search and multi-tenancy isolation.
