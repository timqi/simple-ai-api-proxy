# Makefile for building the executable binary

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=simple-ai-api-proxy

# Build the project
all: build

build:
	$(GOBUILD) -o ./bin/$(BINARY_NAME) -v main.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

test:
	$(GOTEST) -v ./...

.PHONY: all build clean test
