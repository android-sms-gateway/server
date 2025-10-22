project_name = sms-gateway
image_name = capcom6/$(project_name):latest

extension=
ifeq ($(OS),Windows_NT)
	extension = .exe
endif

# Default target
all: fmt lint test benchmark

fmt:
	golangci-lint fmt

# Lint the code using golangci-lint
lint:
	golangci-lint run --timeout=5m

# Run tests with coverage
test:
	go test -race -shuffle=on -count=1 -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...

# Run benchmarks
benchmark:
	go test -run=^$$ -bench=. -benchmem ./... | tee benchmark.txt

# Download dependencies
deps:
	go mod download

# Clean up generated files
clean:
	go clean -cache -testcache
	rm -f coverage.out benchmark.txt

###

init: deps

init-dev: init
	go install github.com/air-verse/air@latest \
		&& go install github.com/swaggo/swag/cmd/swag@latest \
		&& go install github.com/pressly/goose/v3/cmd/goose@latest

ngrok:
	ngrok http 3000

air:
	air

db-upgrade:
	go run ./cmd/$(project_name)/main.go db:migrate

db-upgrade-raw:
	go run ./cmd/$(project_name)/main.go db:auto-migrate
	
run:
	go run cmd/$(project_name)/main.go

test-e2e: test
	cd test/e2e && go test -count=1 .

build:
	go build -o tmp/$(project_name) ./cmd/$(project_name)
	
install:
	go install ./cmd/$(project_name)

docker-build:
	docker build -f build/package/Dockerfile -t $(image_name) --build-arg APP=$(project_name) .

docker:
	docker compose -f deployments/docker-compose/docker-compose.yml up --build

docker-dev:
	docker compose -f deployments/docker-compose/docker-compose.dev.yml up --build

docker-clean:
	docker compose -f deployments/docker-compose/docker-compose.yml down --volumes

.PHONY: all fmt lint test benchmark deps clean init init-dev air ngrok db-upgrade db-upgrade-raw run test-e2e build install docker-build docker docker-dev docker-clean
