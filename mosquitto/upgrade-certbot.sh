#!/bin/sh

touch /ac/upgrade-certbot.log

# https://techjogging.com/create-letsencrypt-certificate-alpine-nginx.html

echo "Calling 'apk update certbot certbot-nginx openssl'..." >> /ac/upgrade-certbot.log
apk update certbot certbot-nginx openssl
echo "Certbot updated!" >> /ac/upgrade-certbot.log
