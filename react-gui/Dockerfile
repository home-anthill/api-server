# syntax=docker/dockerfile:1

FROM node:16-alpine as builder

WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY package*.json ./

RUN npm ci

COPY . ./

RUN npm run build:prod

FROM nginx:1.21.6-alpine

# install certbot with nginx module and openssl (all required to enable certbot with nginx)
# install openrc (contains 'rc-service') and busybox-initscripts to use rc-service crontab
# if you don't add 'busybox-initscripts' you'll get error "rc-update: service `crond' does not exist"
RUN apk update && apk add --no-cache certbot certbot-nginx openssl openrc busybox-initscripts

WORKDIR /ac

COPY --from=builder /app/run.sh run.sh
COPY --from=builder /app/build/ /etc/nginx/html/html

# scripts to renew let's encrypt certs and to upgrade certbot automatically
COPY --from=builder /app/certbot-renew.sh /etc/periodic/daily/certbot-renew.sh
COPY --from=builder /app/upgrade-certbot.sh /etc/periodic/monthly/upgrade-certbot.sh

RUN rm /etc/nginx/conf.d/default.conf

ENTRYPOINT ["sh", "run.sh"]