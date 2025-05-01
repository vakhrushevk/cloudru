LOCAL_BIN:=$(CURDIR)/bin

build:
	go build -o $(LOCAL_BIN)/main ./cmd/main/main.go

run:
	go run ./cmd/main/main.go

docker-build:
	docker-compose build
docker-up:
	docker-compose up -d
docker-rebuild:
	docker-compose up -d --build

docker-logs:
	docker-compose logs -f
docker-logs-main:
	docker-compose logs -f main
docker-logs-backend:
	docker-compose logs -f backend

lint:
	golangci-lint run ./... --config .golangci.yaml
