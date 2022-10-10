# syntax=docker/dockerfile:1
FROM lukemathwalker/cargo-chef:latest-rust-alpine AS chef
WORKDIR /app


FROM chef AS planner
COPY . .
RUN cargo chef prepare --recipe-path recipe.json


FROM chef AS builder
WORKDIR /app

COPY --from=planner /app/recipe.json recipe.json

# Build dependencies - this is the caching Docker layer!
RUN cargo chef cook --release --recipe-path recipe.json

# copy source code
COPY . .
RUN cargo build --release


FROM alpine:latest as runtime
WORKDIR /app

# to run the binary file you need:
# - environment file
# - rocket config file
COPY --from=builder /app/.env .env
COPY --from=builder /app/Rocket.toml Rocket.toml
COPY --from=builder /app/target/release/register register

ENTRYPOINT ["./register"]