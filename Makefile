IMAGE ?= backrest-volsync-operator:dev

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: generate
generate:
	@echo "No code generation configured (CRDs are tracked in config/)."

.PHONY: lint
lint:
	@echo "Lint is optional. Install golangci-lint to enable.";
	@golangci-lint run ./... 2>nul || true

.PHONY: test
test:
	go test ./...

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) .
