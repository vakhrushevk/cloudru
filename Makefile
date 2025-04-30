run:
	go run cmd/main.go

install-golangci-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
lint:
	golangci-lint run ./... --config .golangci.pipeline.yaml
