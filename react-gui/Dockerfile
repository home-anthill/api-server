# syntax=docker/dockerfile:1

FROM node:16-alpine as builder

WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY package*.json ./

RUN npm ci

COPY . ./

RUN npm run build:prod

FROM nginx:alpine
WORKDIR /
COPY --from=builder /app/nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/build/ /etc/nginx/html/html

RUN rm /etc/nginx/conf.d/default.conf

EXPOSE 80 443 8085