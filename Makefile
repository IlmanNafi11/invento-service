.PHONY: help build run test coverage lint clean docker-build docker-run fmt migrate seed install-tools

BINARY_NAME=app
GO=go
DOCKER=docker
DOCKER_IMAGE=invento-backend
DOCKER_TAG=latest

help:
	@echo "Development - Available Commands"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build              - Build the application binary"
	@echo "  make run                - Run the application"
	@echo "  make dev                - Run with hot reload (requires air)"
	@echo ""
	@echo "Testing & Coverage:"
	@echo "  make test               - Run all tests"
	@echo "  make test-v             - Run tests with verbose output"
	@echo "  make coverage           - Run tests with coverage report"
	@echo "  make coverage-html      - Generate HTML coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint               - Run golangci-lint"
	@echo "  make fmt                - Format code with gofmt"
	@echo "  make fmt-check          - Check code formatting"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-run         - Run Docker container"
	@echo "  make docker-push        - Push Docker image to registry"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up         - Run database migrations"
	@echo "  make migrate-down       - Rollback database migrations"
	@echo "  make seed               - Run database seeder"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make clean-all          - Remove all generated files"
	@echo ""
	@echo "Dependencies:"
	@echo "  make install-tools      - Install development tools"
	@echo "  make deps               - Download Go dependencies"
	@echo "  make tidy               - Tidy Go modules"

build:
	@echo "Building application..."
	@mkdir -p bin
	$(GO) build -v -o bin/$(BINARY_NAME) cmd/app/main.go
	@echo "Build successful: bin/$(BINARY_NAME)"

run: build
	@echo "Running application..."
	./bin/$(BINARY_NAME)

dev:
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/cosmtrek/air@latest; }
	@echo "Running with hot reload..."
	air

test:
	@echo "Running tests..."
	$(GO) test -v ./...

test-v:
	@echo "Running tests (verbose)..."
	$(GO) test -v -race ./...

coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -covermode=atomic -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "Tip: Run 'make coverage-html' to view detailed coverage report"

coverage-html: coverage
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Report generated: coverage.html"

lint:
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run --config=.golangci.yml ./...

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted"

fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "Code is not formatted:"; \
		gofmt -s -d .; \
		exit 1; \
	fi
	@echo "Code is properly formatted"

docker-build:
	@echo "Building Docker image..."
	$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	$(DOCKER) build -t $(DOCKER_IMAGE):latest .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: docker-build
	@echo "Running Docker container..."
	$(DOCKER) run -d \
		--name $(DOCKER_IMAGE) \
		-p 3000:3000 \
		--env-file .env \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Container running on port 3000"

docker-stop:
	@echo "Stopping Docker container..."
	$(DOCKER) stop $(DOCKER_IMAGE) || true
	$(DOCKER) rm $(DOCKER_IMAGE) || true
	@echo "Container stopped"

docker-push:
	@echo "Pushing Docker image..."
	$(DOCKER) tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	$(DOCKER) push $(DOCKER_IMAGE):$(DOCKER_TAG)
	$(DOCKER) push $(DOCKER_IMAGE):latest
	@echo "Image pushed to registry"

migrate-up:
	@echo "Running migrations..."
	$(GO) run -tags migrate github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations/app \
		-database "mysql://$${DB_USER}:$${DB_PASSWORD}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_NAME}" \
		up

migrate-down:
	@echo "Rolling back migrations..."
	$(GO) run -tags migrate github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations/app \
		-database "mysql://$${DB_USER}:$${DB_PASSWORD}@tcp($${DB_HOST}:$${DB_PORT})/$${DB_NAME}" \
		down

seed:
	@echo "Running seeder..."
	./bin/$(BINARY_NAME)

clean:
	@echo "Cleaning build artifacts..."
	@rm -f bin/$(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

clean-all: clean
	@echo "Cleaning all generated files..."
	@rm -rf bin/
	@rm -rf vendor/
	@$(GO) clean -cache -testcache
	@echo "All cleaned"

install-tools:
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/cosmtrek/air@latest
	$(GO) install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed"

deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded"

tidy:
	@echo "Tidying Go modules..."
	$(GO) mod tidy
	@echo "Go modules tidy"

verify:
	@echo "Verifying build..."
	@make fmt-check
	@make lint
	@make test
	@echo "All verifications passed"

ci: clean install-tools deps tidy verify build
	@echo "CI workflow complete - ready for production"
