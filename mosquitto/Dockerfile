# syntax=docker/dockerfile:1

FROM eclipse-mosquitto:2.0-openssl

# install certbot and openssl (all required to enable certbot with standalone mode)
# install openrc (contains 'rc-service') and busybox-initscripts to use rc-service crontab
# if you don't add 'busybox-initscripts' you'll get error "rc-update: service `crond' does not exist"
RUN apk update && apk add --no-cache certbot openssl openrc busybox-initscripts

WORKDIR /ac

COPY run.sh run.sh
COPY mosquitto.conf /mosquitto/config/mosquitto.conf

ENTRYPOINT ["sh", "run.sh"]