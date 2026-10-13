APP_DIR := $(shell pwd)

all: menreiki

menreiki:
	@mkdir -p $(APP_DIR)/build
	cd cmd/menreiki && go build -o $(APP_DIR)/build

.PHONY: clean dev

clean: clean-deps
	rm -rf $(APP_DIR)/build

clean-deps:
	go clean -cache
