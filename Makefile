BUILD_DIR := ./build
BINARY := ctrz

NATIVE_OS := $(shell go env GOOS)
NATIVE_ARCH := $(shell go env GOARCH)

TARGETS := \
	linux/amd64 \
	linux/arm64 \
	linux/386 \
	linux/arm

native:
	@mkdir -p $(BUILD_DIR)
	@echo "Building native binary"
	@GOOS=$(NATIVE_OS) GOARCH=$(NATIVE_ARCH) \
	go build -o $(BUILD_DIR)/$(BINARY)-$(NATIVE_OS)-$(NATIVE_ARCH) .

release:
	@mkdir -p $(BUILD_DIR)

	@for target in $(TARGETS); do \
		os=$$(echo $$target | cut -d/ -f1); \
		arch=$$(echo $$target | cut -d/ -f2); \
		echo "Building $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch \
		go build -o $(BUILD_DIR)/$(BINARY)-$$os-$$arch .;\
	done

clean:
	rm -rf $(BUILD_DIR)

.PHONY: native release clean