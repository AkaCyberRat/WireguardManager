.SILENT:


b_api:
	go mod download && mkdir -p ./.bin/api && go build -o ./.bin/api ./cmd/api/main.go


r_api: build_api


b_exp: 
	go mod download && mkdir -p ./.bin/experiments && go build -o ./.bin/experiments ./cmd/experiments/main.go

r_exp: b_exp
	./.bin/experiments/main

