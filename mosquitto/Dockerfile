# syntax=docker/dockerfile:1

FROM eclipse-mosquitto:2.0-openssl

# install certbot and openssl (all required to enable certbot with standalone mode)
# install openrc (contains 'rc-service') and busybox-initscripts to use rc-service crontab
# if you don't add 'busybox-initscripts' you'll get error "rc-update: service `crond' does not exist"
RUN apk update && apk add --no-cache certbot openssl openrc busybox-initscripts

WORKDIR /ac

COPY run.sh run.sh
COPY certbot-renew.sh /etc/periodic/daily/certbot-renew.sh
COPY upgrade-certbot.sh /etc/periodic/monthly/upgrade-certbot.sh

COPY mosquitto.conf /mosquitto/config/mosquitto.conf

# to fix errors with crond
# https://github.com/gliderlabs/docker-alpine/issues/437#issuecomment-494200575
VOLUME ["/sys/fs/cgroup"]

ENTRYPOINT ["sh", "run.sh"]