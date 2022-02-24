# syntax=docker/dockerfile:1
FROM golang:1.17 as builder
WORKDIR /app
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY . ./

RUN go build -o /api-server .

FROM golang:1.17
WORKDIR /
COPY --from=builder /api-server /api-server
COPY --from=builder /app/credentials.json /credentials.json
COPY --from=builder /app/.env_prod /.env_prod

EXPOSE 50051 80 27017 5672

ENTRYPOINT ["/api-server"]