#!/bin/sh
set -e

# Let's encrypt certificates
#Certificate is saved at: /etc/letsencrypt/live/<DOMAIN>/fullchain.pem
#Key is saved at:         /etc/letsencrypt/live/<DOMAIN>/privkey.pem

LOG_FILE=/etc/letsencrypt/certbot-cronjob.log

touch ${LOG_FILE}

if [ -d "/etc/letsencrypt/live/${CERTBOT_DOMAIN}" ]
then
  echo "$(date) - Certificates already exists. Try to renew it via certbot." >> ${LOG_FILE}
  # https://techjogging.com/create-letsencrypt-certificate-alpine-nginx.html
  echo "$(date) - Calling 'certbot renew'..." >> ${LOG_FILE}
  certbot renew >> ${LOG_FILE}
  echo "$(date) - Certbot renew executed!" >> ${LOG_FILE}
else
  echo "$(date) - Certificates not available. Nothing to renew."
fi