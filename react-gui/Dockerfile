# syntax=docker/dockerfile:1

FROM node:16-alpine as builder

WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY package*.json ./

RUN npm ci

COPY . ./

RUN npm run build:prod

FROM nginx:1.21.6-alpine
# install certbot and openrc (container 'rc-service')
RUN apk update --no-cache add certbot certbot-nginx openssl openrc
WORKDIR /
COPY --from=builder /app/build/ /etc/nginx/html/html
COPY --from=builder /app/run.sh /run.sh

RUN rm /etc/nginx/conf.d/default.conf

EXPOSE 80 443 8085

ENTRYPOINT ["sh", "/run.sh"]