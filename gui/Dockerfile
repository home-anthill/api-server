# syntax=docker/dockerfile:1

FROM node:16-alpine as builder

WORKDIR /ac

ENV PATH /ac/node_modules/.bin:$PATH

COPY package.json .
COPY package-lock.json .

RUN npm ci

COPY . .

RUN npm run build:prod

FROM nginx:1.21.6-alpine

# install certbot with nginx module and openssl (all required to enable certbot with nginx)
# install openrc (contains 'rc-service') and busybox-initscripts to use rc-service crontab
# if you don't add 'busybox-initscripts' you'll get error "rc-update: service `crond' does not exist"
RUN apk update && apk add --no-cache \
    certbot certbot-nginx openssl \
    openrc busybox-initscripts

WORKDIR /ac

COPY --from=builder /ac/run.sh run.sh
COPY --from=builder /ac/build/ /etc/nginx/html/html

# scripts to renew let's encrypt certs and to upgrade certbot automatically.
# Filenames in `/etc/periodic` must follow specific rules: https://askubuntu.com/a/611430
# For example you cannot use "dot" in filenames, so you must remove file extensions
COPY --from=builder /ac/certbot-renew.sh /etc/periodic/daily/certbot-renew
COPY --from=builder /ac/upgrade-certbot.sh /etc/periodic/monthly/upgrade-certbot

# remove nginx default configuration
RUN rm /etc/nginx/conf.d/default.conf

# to fix errors with crond
# https://github.com/gliderlabs/docker-alpine/issues/437#issuecomment-494200575
VOLUME ["/sys/fs/cgroup"]

ENTRYPOINT ["sh", "run.sh"]