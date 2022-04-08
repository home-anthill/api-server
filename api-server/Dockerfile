# syntax=docker/dockerfile:1
FROM golang:1.17-alpine as builder
RUN apk update && apk add --no-cache curl make unzip protoc
WORKDIR /app
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY . ./

#RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip
#RUN unzip protoc-3.15.8-linux-x86_64.zip -d $HOME/.local
#RUN export PATH="$PATH:$HOME/.local/bin"
#
#RUN make proto
#RUN protoc api/gRPC/*/*.proto \
#    			--go_out=. \
#    			--go_opt=paths=source_relative \
#    			--go-grpc_out=. \
#    			--go-grpc_opt=paths=source_relative \
#    			--proto_path=.

RUN go build -o /api-server .

FROM golang:1.17-alpine
WORKDIR /
COPY --from=builder /api-server /api-server
COPY --from=builder /app/credentials.json /credentials.json
COPY --from=builder /app/cert /cert

EXPOSE 80 443

ENTRYPOINT ["/api-server"]