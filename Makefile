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



lint:
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



lint-fix:
	docker run -it \
		--rm \
		-v "$(CURDIR):/app" \
		-v golangci-lint-cache:/root/.cache \
		-v go-mod-cache:/go/pkg/mod \
		-w /app \
		--name golangci-lint-container \
		golangci/golangci-lint:v1.63.4 \
		bash -c 'start=$$(date +%s); \
			echo "Starting golangci-lint --fix ..."; \
			golangci-lint run ./... --config .golangci.yml --fix; \
			end=$$(date +%s); \
			elapsed=$$((end - start)); \
			echo "Lint fix completed! Time: $$elapsed sec"'

