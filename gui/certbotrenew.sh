#!/bin/sh

touch /ac/certbotrenew.log

# https://techjogging.com/create-letsencrypt-certificate-alpine-nginx.html

echo "Calling 'certbot renew'..." >> /ac/certbotrenew.log
certbot renew
echo "Calling renew done!" >> /ac/certbotrenew.log