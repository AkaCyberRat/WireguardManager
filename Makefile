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
	golangci-lint run ./... --config .golangci.yml; \
	echo '"make lint-check" completed!'; \
	end=$$(date +%s); \
	elapsed=$$((end - start)); \
	echo "Time: $$elapsed sec"


lint-fix:
	@start=$$(date +%s); \
	golangci-lint run ./... --config .golangci.yml --fix; \
	echo '"make lint-fix" completed!'; \
	end=$$(date +%s); \
	elapsed=$$((end - start)); \
	echo "Time: $$elapsed sec"