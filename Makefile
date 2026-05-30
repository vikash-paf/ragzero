.PHONY: build run up down test lint migrate-new clean

# Build the Go application
build:
	go build -o bin/api cmd/api/main.go

# Run the Go application locally (ensure dependencies are up)
run:
	go run cmd/api/main.go

# Start all services via Docker Compose
up:
	docker-compose up -d --build

# Stop and remove all Docker containers
down:
	docker-compose down -v

# Run tests
test:
	go test -v ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Create a new migration file
# Usage: make migrate-new name=your_migration_name
migrate-new:
	@read -p "Enter migration name: " name; \
	VERSION=$$(date +%s); \
	touch migrations/$${VERSION}_$${name}.up.sql; \
	touch migrations/$${VERSION}_$${name}.down.sql; \
	echo "Created migrations/$${VERSION}_$${name}.up.sql and .down.sql"

# Clean build artifacts
clean:
	rm -rf bin/
