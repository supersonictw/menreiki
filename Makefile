APP_DIR := $(shell pwd)

all: menreiki

menreiki:
	@mkdir -p $(APP_DIR)/build
	cd cmd/menreiki && go build -o $(APP_DIR)/build

amd64:
	@mkdir -p $(APP_DIR)/build
	podman run --rm -v $(APP_DIR):/app -w /app/cmd/menreiki golang:alpine sh -c "\
		apk add --no-cache zig musl-dev && \
		CC='zig cc -target x86_64-linux-musl' CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o /app/build/menreiki-linux-amd64 \
	"

arm64:
	@mkdir -p $(APP_DIR)/build
	podman run --rm -v $(APP_DIR):/app -w /app/cmd/menreiki golang:alpine sh -c "\
		apk add --no-cache zig musl-dev && \
		CC='zig cc -target aarch64-linux-musl' CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags='-s -w' -o /app/build/menreiki-linux-arm64 \
	"

cross: amd64 arm64

.PHONY: clean dev amd64 arm64 cross

clean: clean-deps
	rm -rf $(APP_DIR)/build

clean-deps:
	go clean -cache
