build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/collie-agent-amd64 .
	docker build -t collie.azurecr.io/collie-agent:1 .

generate:
	go generate ./...

push:
	docker push collie.azurecr.io/collie-agent:1

deploy:
	cat deployment.yaml | envsubst | kubectl apply -f -

SHELL := /bin/bash
run:
	source ./.env && go run .

test:
	go test ./... -race

lint:
	go vet ./...
	gofmt -w -s .
	golangci-lint run

release: build push
