.PHONY: test
test:
	docker run -t --rm -v $(PWD):/app -w /app golang:1.24.4-alpine3.22 go test -v ./...

.PHONY: lint
lint:
	docker run -t --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.64.8 golangci-lint run
