# Ragzero: Distributed Document Search Service

Ragzero is a high-performance, multi-tenant document search service built with Go 1.26, Elasticsearch, Redis, and PostgreSQL. It is designed to handle millions of documents with sub-500ms latency.

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

### 2. API Documentation & Testing
Sample API requests and lifecycle tests are available as `.http` files in the following directory:
- [docs/http/](./docs/http/)

These files can be executed directly within GoLand or VS Code (with the REST Client extension) to verify system functionality.

## Production Readiness
For a detailed analysis on scaling to 100x, resilience, and security, please refer to the architecture documentation (stored in `.tmp/architecture.md` for this assessment).
