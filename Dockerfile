# syntax=docker/dockerfile:1

# ── Stage 1: Builder ────────────────────────────────────────────────
FROM golang:1.26.3-alpine AS builder
RUN apk update && apk add --no-cache \
    protoc \
    make gcc musl-dev

# install protoc requirements based on https://grpc.io/docs/languages/go/quickstart/
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
ENV PATH="$PATH:$(go env GOPATH)/bin"

WORKDIR /app
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY . ./

RUN make deps

RUN make build

# Pre-create a dedicated, empty logs directory owned by nobody (uid/gid 65534)
# We do this in a separate folder so we don't accidentally copy source code
RUN mkdir -p /scratch/logs && chown -R 65534:65534 /scratch/logs

# ── Stage 2: Hardened runtime ────────────────────────────────────────────────
FROM dhi.io/alpine-base:3.23

WORKDIR /

# App directory skeleton (empty /logs owned by nobody).
COPY --from=builder --chown=65534:65534 /scratch/logs /logs

# Binary and env template.
COPY --from=builder --chown=65534:65534 /app/build/api-server /api-server

USER 65534

ENTRYPOINT ["/api-server"]
