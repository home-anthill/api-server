# syntax=docker/dockerfile:1
FROM golang:1.17 as builder
WORKDIR /app
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY . ./

RUN go build -o /api-devices .

FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=builder /api-devices /api-devices
COPY --from=builder /app/.env_prod /.env_prod

EXPOSE 50051 27017 1883 5672

ENV ENV="prod"

ENTRYPOINT ["/api-devices"]