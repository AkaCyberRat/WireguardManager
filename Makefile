.SILENT:



api_build_local:
	go mod download && mkdir -p ./.bin/api && go build -o ./.bin/api ./cmd/api/main.go

api_run_local: api_build_local
	./.bin/api/main

api_run_docker:
	docker-compose -f ./deploy/docker-compose.yml up --build

