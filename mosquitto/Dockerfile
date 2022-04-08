# syntax=docker/dockerfile:1

FROM eclipse-mosquitto:2.0-openssl

# install certbot and openrc (container 'rc-service')
RUN apk update && apk add --no-cache certbot openssl openrc

WORKDIR /ac

COPY run.sh run.sh
COPY mosquitto.conf /mosquitto/config/mosquitto.conf

ENTRYPOINT ["sh", "run.sh"]