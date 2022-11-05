#!/bin/sh

touch /ac/upgradecertbot.log

# https://techjogging.com/create-letsencrypt-certificate-alpine-nginx.html

echo "Calling 'apk update certbot certbot-nginx openssl'..." >> /ac/upgradecertbot.log
apk update certbot certbot-nginx openssl
echo "Certbot updated!" >> /ac/upgradecertbot.log