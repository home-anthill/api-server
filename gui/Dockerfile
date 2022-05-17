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
RUN apk update && apk add --no-cache \
    certbot certbot-nginx openssl

WORKDIR /ac

COPY --from=builder /ac/run.sh run.sh
COPY --from=builder /ac/build/ /etc/nginx/html/html

# remove nginx default configuration
RUN rm /etc/nginx/conf.d/default.conf

ENTRYPOINT ["sh", "run.sh"]