# syntax=docker/dockerfile:1
FROM eclipse-mosquitto:2.0-openssl

RUN apk --update --no-cache add certbot certbot-nginx openssl

WORKDIR /ac

COPY run.sh run.sh
COPY mosquitto.conf /mosquitto/config/mosquitto.conf

ENTRYPOINT ["sh", "run.sh"]