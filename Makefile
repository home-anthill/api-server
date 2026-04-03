.DEFAULT_GOAL := build

lint:
	set +e
	# use staticcheck, because golint has been deprecated
	staticcheck ./...
	set -e
.PHONY:lint

vet:
	go vet ./...
	shadow ./...
.PHONY:vet

check:
	# find vulnerabilities
	govulncheck ./...
.PHONY:check

proto:
	protoc api/grpc/*/*.proto \
			--go_out=. \
			--go_opt=paths=source_relative \
			--go-grpc_out=. \
			--go-grpc_opt=paths=source_relative \
			--proto_path=.
.PHONY: proto

build: proto vet lint
	go build -o ./build/api-server .
.PHONY: build

run: proto vet lint
	air
.PHONY: run

test: proto vet lint
	mkdir -p ./coverage
	ENV=testing go test -v -race -count=1 -coverpkg ./... -coverprofile ./coverage/profile.cov ./...
	# go tool cover -html ./coverage/profile.cov
	go tool cover -html ./coverage/profile.cov -o ./coverage/cover.html
	go-cover-treemap -coverprofile ./coverage/profile.cov > ./coverage/out.svg
.PHONY: test

deps:
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/air-verse/air@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go get -u
	go mod tidy
	go install github.com/nikolaydubina/go-cover-treemap@latest
	echo "You should update protoc manually: https://grpc.io/docs/protoc-installation/"
.PHONY: deps
