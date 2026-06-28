BUILD_DIR := ./build
BINARY := ctrz

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

TARGETS := \
	linux/amd64 \
	linux/arm64 \
	linux/386 \
	linux/arm

VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X 'ctrz/misc.Version=$(VERSION)' \
           -X 'ctrz/misc.Commit=$(COMMIT)' \
           -X 'ctrz/misc.BuildDate=$(DATE)'

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
	go test ctrz/cgroup ctrz/misc ctrz/network ctrz/proc -vet=all

clean:
	rm -rf $(BUILD_DIR)

.PHONY: build release clean test