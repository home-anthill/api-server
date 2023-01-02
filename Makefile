.DEFAULT_GOAL := build

# you cannot customize fmt to don't use tabs,
# so at the moment I disabled this command
#fmt:
#	go fmt ./...
#.PHONY:fmt

lint:
	golint ./...
.PHONY:lint

vet:
	go vet ./...
	shadow ./...
.PHONY:vet

proto:
	protoc api/gRPC/*/*.proto \
			--go_out=. \
			--go_opt=paths=source_relative \
			--go-grpc_out=. \
			--go-grpc_opt=paths=source_relative \
			--proto_path=.
.PHONY: proto

build: vet proto
	go build -o ./build/api-server .
.PHONY: build

run: vet proto
	air
.PHONY: run

test:
	mkdir -p ./coverage
	ENV=testing go test -v -count=1 -coverpkg=./... -coverprofile=./coverage/profile.cov ./...
	go tool cover -html=./coverage/profile.cov
	go-cover-treemap -coverprofile ./coverage/profile.cov > ./coverage/out.svg
.PHONY: test

deps:
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
	go install golang.org/x/tools/cmd/cover
	go install github.com/nikolaydubina/go-cover-treemap@latest
	go get -u
	go mod tidy
.PHONY: deps
