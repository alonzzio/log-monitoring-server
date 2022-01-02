# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=lms
BINARY_UNIX=$(BINARY_NAME)_unix
PROJECT_NAME = $(notdir $(PWD))
SERVICE_TARGET := main

.PHONY: build test clean run dep

all: docker-up dep test build run
build:
		$(GOBUILD) -o $(BINARY_NAME) -v
test:
		$(GOTEST) -v ../...
clean:
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
run:
		$(GOBUILD) -o $(BINARY_NAME) -v ./...
		./$(BINARY_NAME)
dep:
	$(GOMOD) download

#Docker compose up
# Mysql image will pull and run
docker-up:
	docker-compose -f ../docker-compose.yml up -d

docker-kill:
	docker kill mysql-database

docker-clean:
	docker container rm --force mysql-database


