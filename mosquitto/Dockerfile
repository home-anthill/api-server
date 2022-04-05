# syntax=docker/dockerfile:1

FROM eclipse-mosquitto:2.0-openssl

RUN apk update && apk add certbot certbot-nginx openssl

COPY run.sh ./

ENTRYPOINT ["sh", "/run.sh"]