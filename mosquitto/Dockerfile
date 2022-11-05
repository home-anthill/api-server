# syntax=docker/dockerfile:1

FROM eclipse-mosquitto:2.0-openssl

# install certbot and openssl (all required to enable certbot with standalone mode)
# install openrc (contains 'rc-service') and busybox-initscripts to use rc-service crontab
# if you don't add 'busybox-initscripts' you'll get error "rc-update: service `crond' does not exist"
RUN apk update && apk add --no-cache \
    certbot openssl \
    openrc busybox-initscripts

WORKDIR /ac

COPY run.sh run.sh
COPY init-crond.sh init-crond.sh

# Scripts to renew let's encrypt certs and to upgrade certbot automatically.
# Filenames in `/etc/periodic` must follow specific rules: https://askubuntu.com/a/611430
# For example you cannot use "dot" in filenames, so you must remove file extensions
COPY certbot-renew.sh /etc/periodic/daily/certbot-renew
COPY upgrade-certbot.sh /etc/periodic/monthly/upgrade-certbot

COPY mosquitto.conf /mosquitto/config/mosquitto.conf
COPY mosquitto-no-security.conf /mosquitto/config/mosquitto-no-security.conf

# to fix errors with crond
# https://github.com/gliderlabs/docker-alpine/issues/437#issuecomment-494200575
VOLUME ["/sys/fs/cgroup"]

ENTRYPOINT ["sh", "run.sh"]