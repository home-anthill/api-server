#!/bin/sh

touch /ac/certbot-renew.log

# https://techjogging.com/create-letsencrypt-certificate-alpine-nginx.html

echo "Calling 'certbot renew'..." >> /ac/certbot-renew.log
certbot renew
echo "Calling renew done!" >> /ac/certbot-renew.log
