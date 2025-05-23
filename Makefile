.PHONY: build clean test lint docker-build

APP_NAME = azure-sas-token-extractor
DOCKER_IMAGE = azure-sas-token-extractor:latest

build:
	go build -o $(APP_NAME)

clean:
	rm -f $(APP_NAME)

test:
	go test ./... -v

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run --rm $(DOCKER_IMAGE) --help

all: clean build test

.DEFAULT_GOAL := build