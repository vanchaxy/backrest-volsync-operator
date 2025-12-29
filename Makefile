IMAGE ?= backrest-volsync-operator:dev

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: generate
generate:
	@echo "No code generation configured (CRDs are tracked in config/)."

.PHONY: lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "Lint skipped (golangci-lint not installed)."; \
	fi

.PHONY: test
test:
	go test ./...

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) .
