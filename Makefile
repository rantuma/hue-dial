VERSION  ?= dev
OUTPUT   ?= ./bin/hue
IMAGE    ?= ghcr.io/rantuma/hue-dial

BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS    := -s -w -X github.com/rantuma/hue-dial/pkg/version.Version=$(VERSION) -X github.com/rantuma/hue-dial/pkg/version.BuildTime=$(BUILD_TIME)

.PHONY: build test lint licenses publish

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o "$(OUTPUT)" ./main.go
	@echo "Built hue version: $(VERSION) at $(BUILD_TIME)"

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run --timeout=5m

licenses:
	go-licenses check .
	rm -rf licenses
	go-licenses save . --save_path=./licenses

publish:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION="$(VERSION)" \
		--tag "$(IMAGE):$(VERSION)" \
		--tag "$(IMAGE):latest" \
		--push \
		.
	@echo "Published $(IMAGE):$(VERSION) and $(IMAGE):latest"
