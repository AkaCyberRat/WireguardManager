.SILENT:



api_build_local:
	go mod download && mkdir -p ./.bin/api && go build -o ./.bin/api ./cmd/api/main.go

api_run_local: api_build_local
	./.bin/api/main

api_run_docker:
	docker-compose -f ./deploy/docker-compose.yml up --build

api_run_docker_dev:
	docker-compose -f ./deploy/docker-compose.yml.dev up --build



example_build:
	go mod download && mkdir -p ./.bin/example && go build -o ./.bin/example ./cmd/example/main.go

example_run: example_build

	 ./.bin/example/main


lint-check:
	@start=$$(date +%s); \
	docker run --rm \
		-v "$(CURDIR):/app" \
		-v "$(HOME)/go/pkg/mod:/go/pkg/mod" \
		-v "$(HOME)/.cache/golangci-lint:/root/.cache" \
		-w /app \
		golangci/golangci-lint:v1.63.4 \
		golangci-lint run ./... --config .golangci.yml; \
	echo '"make lint-check" completed!'; \
	end=$$(date +%s); \
	elapsed=$$((end - start)); \
	echo "Time: $$elapsed sec"


lint-fix:
	@start=$$(date +%s); \
	docker run --rm \
		-v "$(CURDIR):/app" \
		-v "$(HOME)/go/pkg/mod:/go/pkg/mod" \
		-v "$(HOME)/.cache/golangci-lint:/root/.cache" \
		-w /app \
		golangci/golangci-lint:v1.63.4 \
		golangci-lint run ./... --config .golangci.yml --fix; \
	echo '"make lint-fix" completed!'; \
	end=$$(date +%s); \
	elapsed=$$((end - start)); \
	echo "Time: $$elapsed sec"


# ----------------------------
# Конфигурация
# ----------------------------
APP_NAME := single-peer-setup
BIN_DIR := .bin/single-peer-setup
BIN_PATH := $(BIN_DIR)/$(APP_NAME)

CONFIG_SRC := cmd/testing/single-peer-setup/config.json
CONFIG_DST := $(BIN_DIR)/config.json

DOCKER_IMAGE := single-peer-setup-run
DOCKERFILE := cmd/testing/single-peer-setup/Dockerfile

TARGET_OS := linux
TARGET_ARCH := amd64

# ----------------------------
# Public targets
# ----------------------------

.PHONY: run build docker-image clean

## ОДНА КОМАНДА: билд + запуск контейнера
run: build docker-image
	docker run --rm \
		--name $(APP_NAME) \
		--privileged \
		-p 51830:51821/udp \
		-v $(PWD)/$(BIN_DIR):/app \
		$(DOCKER_IMAGE)

## Локальная сборка бинарника под контейнер
build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 \
	GOOS=$(TARGET_OS) \
	GOARCH=$(TARGET_ARCH) \
	go build -o $(BIN_PATH) ./cmd/testing/single-peer-setup/main.go
	cp $(CONFIG_SRC) $(CONFIG_DST)

## Сборка runtime-образа (очень быстрая, почти кешируемая)
docker-image:
	docker build \
		-t $(DOCKER_IMAGE) \
		-f $(DOCKERFILE) .

clean:
	rm -rf $(BIN_DIR)
