# syntax=docker/dockerfile:1
FROM rust:alpine3.16 AS chef

# some cargo dependencies require additional packages
# to build the project.
RUN apk update && apk add --no-cache \
    g++

WORKDIR /app

RUN cargo install cargo-chef


FROM chef AS planner

COPY . .

RUN cargo chef prepare --recipe-path recipe.json


FROM chef AS builder

WORKDIR /app

COPY --from=planner /app/recipe.json recipe.json

# Build dependencies - this is the caching Docker layer!
RUN cargo chef cook --release --recipe-path recipe.json

# copy source code and build it
COPY . .

RUN cargo build --release


FROM alpine:latest as runtime

# to be able to use ROOT CAs file from /etc/ssl/certs/
# folder, you must install the 'ca-certificates' package
RUN apk update && apk add --no-cache \
    ca-certificates

WORKDIR /app

# to run the binary file you need:
# - environment file
# - rocket config file
COPY --from=builder /app/log4rs.yaml log4rs.yaml
COPY --from=builder /app/target/release/consumer consumer

ENTRYPOINT ["./consumer"]