.PHONY: run test swag build docker-build docker-up docker-down

# Run the API locally
run:
	go run cmd/main.go

# Run unit and integration tests
test:
	go test -v ./tests/...

# Run microsecond benchmarks
bench:
	cd tests && go test -bench=. -benchmem -run=^$$

# Regenerate Swagger OpenAPI documentation
swag:
	swag init -g cmd/main.go -o docs

# Compile production binary
build:
	go build -trimpath -ldflags="-s -w" -o bin/expertlisting-server cmd/main.go

# Build Docker image
docker-build:
	docker build -t expertlisting-server .

# Start Docker Compose services
docker-up:
	docker compose up -d

# Stop Docker Compose services
docker-down:
	docker compose down
