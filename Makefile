BUILD_DIR := ./build
BINARY := ctrz

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

TARGETS := \
	linux/amd64 \
	linux/arm64 \
	linux/386 \
	linux/arm

VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo dev)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X 'ctrz/version.Version=$(VERSION)' \
           -X 'ctrz/version.Commit=$(COMMIT)' \
           -X 'ctrz/version.BuildDate=$(DATE)'

build:
	@echo "Building binary for GOOS=$(GOOS) and GOARCH=$(GOARCH)"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build -ldflags "$(LDFLAGS)" -o $(BINARY)-$(GOOS)-$(GOARCH) .

release: test
	@mkdir -p $(BUILD_DIR)

	@for target in $(TARGETS); do \
		os=$$(echo $$target | cut -d/ -f1); \
		arch=$$(echo $$target | cut -d/ -f2); \
		echo "Building $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch \
		go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-$$os-$$arch .;\
	done

test:
	go test ctrz/cgroup ctrz/logging ctrz/runtime ctrz/network ctrz/proc ctrz/fs -vet=all

clean:
	rm -rf $(BUILD_DIR)

.PHONY: build release clean test