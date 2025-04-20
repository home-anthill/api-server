# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
RUN apk update && apk add --no-cache \
    protoc \
    make gcc musl-dev

# install protoc requirements based on https://grpc.io/docs/languages/go/quickstart/
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
ENV PATH "$PATH:$(go env GOPATH)/bin"

WORKDIR /app
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY . ./

RUN make deps

RUN make build

FROM golang:1.24-alpine
WORKDIR /
COPY --from=builder /app/build/api-server /api-server
COPY --from=builder /app/.env_template /.env

ENTRYPOINT ["/api-server"]
