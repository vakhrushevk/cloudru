LOCAL_BIN:=$(CURDIR)/bin

run:
	go run cmd/main.go

docker-build:
	docker-compose build
docker-up:
	docker-compose up -d
docker-rebuild:
	docker-compose up -d --build

docker-logs:
	docker-compose logs -f
docker-logs-app:
	docker-compose logs -f app


install-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3

lint:
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.yaml
