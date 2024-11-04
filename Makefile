#!make

run:
	@echo "Running server..."
	@go run main.go

test:
	@echo "Running tests..."
	@go test -cover ./...

lint:
	@echo "Running linters..."
	@golangci-lint run

format:
	@echo "Running formating code..."
	@golangci-lint run --fix
