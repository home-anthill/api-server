# syntax=docker/dockerfile:1

FROM alpine:latest

# install certbot and openssl (all required to enable certbot with standalone mode)
RUN apk update && apk add --no-cache \
    certbot openssl

WORKDIR /ac

COPY run.sh run.sh

ENTRYPOINT ["sh", "run.sh"]