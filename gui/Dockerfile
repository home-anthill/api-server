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

# install certbot and openssl (all required to enable certbot with standalone mode)
RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    certbot certbot-nginx openssl

WORKDIR /ac

COPY --from=builder /ac/run.sh run.sh
COPY --from=builder /ac/build/ /etc/nginx/html/html
# Scripts to renew let's encrypt certs and to upgrade certbot automatically.
# Filenames in `/etc/periodic` must follow specific rules: https://askubuntu.com/a/611430
# For example you cannot use "dot" and "dash" in filenames, so you must remove file extensions
COPY --from=builder /ac/certbotrenew.sh /etc/periodic/15min/certbotrenew
# you have to make all files in /etc/periodic executables,
# otherwise crond will skip non-exetubale files without throwing errors
RUN chmod +x /etc/periodic/15min/certbotrenew

# remove nginx default configuration
RUN rm /etc/nginx/conf.d/default.conf

ENTRYPOINT ["sh", "run.sh"]