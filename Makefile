MODULE_NAME=app
VERSION=$(shell git describe --tags --always)
COMMIT=$(shell git rev-parse HEAD)
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS=-X '$(MODULE_NAME)/internal/version.Version=$(VERSION)' \
        -X '$(MODULE_NAME)/internal/version.CommitHex=$(COMMIT)' \
        -X '$(MODULE_NAME)/internal/version.BuildTime=$(BUILD_TIME)'

GO_DIR=/home/dev/Desktop/Go
APP_DIR=$(GO_DIR)/app
BUILD_DIR=$(GO_DIR)/makisama_builds/$(VERSION)
MAIN_DIR=$(APP_DIR)/cmd/web



.PHONY: build
build:
	mkdir -p $(BUILD_DIR)/ && \
	cp -r $(APP_DIR)/migrations $(BUILD_DIR)/ && \
	cp -r $(APP_DIR)/static $(BUILD_DIR)/ && \
	cp -r $(APP_DIR)/templates $(BUILD_DIR)/ && \
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/app $(MAIN_DIR)/main.go

.PHONY: run
run:
	go run -ldflags "$(LDFLAGS)" $(MAIN_DIR)/main.go