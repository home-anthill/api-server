# syntax=docker/dockerfile:1

FROM alpine:latest

# install certbot and openssl (all required to enable certbot with standalone mode)
RUN apk update && apk add --no-cache \
    certbot openssl

WORKDIR /ac

COPY run.sh run.sh

# Scripts to renew let's encrypt certs and to upgrade certbot automatically.
# Filenames in `/etc/periodic` must follow specific rules: https://askubuntu.com/a/611430
# For example you cannot use "dot" in filenames, so you must remove file extensions
COPY certbot-renew.sh /etc/periodic/daily/certbot-renew
COPY upgrade-certbot.sh /etc/periodic/monthly/upgrade-certbot

ENTRYPOINT ["sh", "run.sh"]