.SILENT:



api_build_local:
	go mod download && mkdir -p ./.bin/api && go build -o ./.bin/api ./cmd/api/main.go

api_run_local: api_build_local
	./.bin/api/main

api_run_docker:
	docker-compose -f ./deploy/docker-compose.yml up --build

api_run_docker_dev:
	docker-compose -f ./deploy/docker-compose.yml.dev up --build


 
exmpl_build: 
	go mod download && mkdir -p ./.bin/example && go build -o ./.bin/example ./cmd/example/main.go

exmpl_run: exmpl_build
	
	 ./.bin/example/main


lint-check:
	golangci-lint run ./... --config .golangci.yml


lint-fix:
	golangci-lint run ./... --config .golangci.yml --fix