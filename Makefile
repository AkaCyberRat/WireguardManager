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


lint-check-windows:
	docker run -it \
		--rm \
		-v "$(CURDIR):/app" \
		-v golangci-lint-cache:/root/.cache \
		-v go-mod-cache:/go/pkg/mod \
		-w /app \
		--name golangci-lint-container \
		golangci/golangci-lint:v1.63.4 \
		bash -c 'start=$$(date +%s); \
			echo "Starting golangci-lint..."; \
			golangci-lint run ./... --config .golangci.yml; \
			end=$$(date +%s); \
			elapsed=$$((end - start)); \
			echo "Lint completed! Time: $$elapsed sec"'