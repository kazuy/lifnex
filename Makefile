IMAGE_URI ?= lifnex:local
PLATFORM ?= linux/amd64
GO_APP_DIR ?= app
GO_BINARY ?= ./bin/lifnex

.PHONY: build-image go-format-check go-vet go-test go-build

build-image:
	docker build --platform $(PLATFORM) --tag $(IMAGE_URI) app

go-format-check:
	@unformatted="$$(gofmt -l $(GO_APP_DIR))"; \
	if [ -n "$${unformatted}" ]; then \
		printf 'The following files are not formatted:\n%s\n' "$${unformatted}"; \
		exit 1; \
	fi

go-vet:
	go -C $(GO_APP_DIR) vet ./...

go-test:
	go -C $(GO_APP_DIR) test ./...

go-build:
	mkdir -p $(GO_APP_DIR)/bin
	go -C $(GO_APP_DIR) build -o $(GO_BINARY) ./cmd/lifnex
