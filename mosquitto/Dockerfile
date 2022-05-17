# syntax=docker/dockerfile:1

FROM eclipse-mosquitto:2.0-openssl

# install certbot and openssl (all required to enable certbot with standalone mode)
RUN apk update && apk add --no-cache \
    certbot openssl

WORKDIR /ac

COPY run.sh run.sh

COPY mosquitto.conf /mosquitto/config/mosquitto.conf
COPY mosquitto-no-security.conf /mosquitto/config/mosquitto-no-security.conf

ENTRYPOINT ["sh", "run.sh"]