#!/bin/sh
set -e

# Let's encrypt certificates
#Certificate is saved at: /etc/letsencrypt/live/<DOMAIN>/fullchain.pem
#Key is saved at:         /etc/letsencrypt/live/<DOMAIN>/privkey.pem

# read env variables
echo "Env variables:"
echo "CERTBOT_DOMAIN = ${CERTBOT_DOMAIN}"

ls /etc/letsencrypt

# define and create logfile
LOG_FILE=/etc/letsencrypt/certbotrenew.log
echo ${LOG_FILE}
touch ${LOG_FILE}


if [ ${CERTBOT_DOMAIN+x} ] && [ -d "/etc/letsencrypt/live/${CERTBOT_DOMAIN}" ]; then
  echo "$(date) - Certificates already exists!" >> ${LOG_FILE}

  echo "$(date) - Calling 'apk update certbot openssl'..." >> ${LOG_FILE}
  # update certbot to use always the latest stable version
  apk update certbot openssl >> ${LOG_FILE}
  echo "$(date) - Certbot updated!" >> ${LOG_FILE}

  # https://techjogging.com/create-letsencrypt-certificate-alpine-nginx.html
  echo "$(date) - Calling 'certbot renew'..." >> ${LOG_FILE}
  # renew certificates (if required) and send the result to the logfile
  certbot renew >> ${LOG_FILE}
  echo "$(date) - Certbot renew executed!" >> ${LOG_FILE}
else
  echo "$(date) - Certificates not available. Nothing to renew."
fi