build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/collie-api-server-amd64 .
	docker rmi --force collie.azurecr.io/collie-api-server:1
	docker build -t collie.azurecr.io/collie-api-server:1 .

swagger:
	go get -u github.com/swaggo/swag/cmd/swag
	swag init

generate:
	go generate ./...

push:
	docker push collie.azurecr.io/collie-api-server:1

deploy:
	cat ../helm-charts/api-server.yaml | envsubst | kubectl apply -f -

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
