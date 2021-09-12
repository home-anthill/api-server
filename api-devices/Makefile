.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:fmt

vet: fmt
	go vet ./...
	shadow ./...
.PHONY:vet

build: vet
	go build -o air-conditioner main.go
.PHONY: build