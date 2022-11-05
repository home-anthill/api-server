#!/bin/sh
set -e

if [ ${CERTBOT_EMAIL+x} ] && [ ${CERTBOT_DOMAIN+x} ] && [ ${CERTBOT_SERVER+x} ]; then
  # Let's encrypt certificates
  #Certificate is saved at: /etc/letsencrypt/live/<DOMAIN>/fullchain.pem
  #Key is saved at:         /etc/letsencrypt/live/<DOMAIN>/privkey.pem

  # read env variables
  echo "Env variables:"
  echo "CERTBOT_EMAIL = ${CERTBOT_EMAIL}"
  echo "CERTBOT_DOMAIN = ${CERTBOT_DOMAIN}"
  echo "CERTBOT_SERVER = ${CERTBOT_SERVER}"

  # https://techjogging.com/create-letsencrypt-certificate-alpine-nginx.html
  echo "Checking 'crond' existence"
  rc-service --list | grep -i crond
  echo "Checking running services"
  rc-status -a
  sleep 5
  # add syslog (required by crontab)
  echo "Enabling syslog"
  rc-update add syslog boot
  sleep 2
  echo "Preparing syslog"
  touch /run/openrc/softlevel
  echo "Starting syslog"
  rc-service syslog start
  sleep 5
  # start crontab
  echo "Starting crond"
  rc-update add crond default
  sleep 2
  rc-service crond start
  sleep 5
  # check services
  echo "Checking running services again"
  rc-status -a
  crontab -l
  # to read log messages
  # tail -f /var/log/messages

  if [ -d "/etc/letsencrypt/live/${CERTBOT_DOMAIN}" ]
  then
    echo "Certificates already exists"
  else
    echo "Requesting certificates using certbot via production server"
    certbot certonly --standalone -m "${CERTBOT_EMAIL}" --agree-tos -d ${CERTBOT_DOMAIN},www.${CERTBOT_DOMAIN} -n --server "${CERTBOT_SERVER}"
    # staging server:
    # certbot certonly --standalone -m "${CERTBOT_EMAIL" --agree-tos -d ${certbot_domain},www.${certbot_domain} -n --server https://acme-staging-v02.api.letsencrypt.org/directory
  fi

  CERT_DIR=/etc/mosquitto/certs
  mkdir -p ${CERT_DIR}

  cp "/etc/letsencrypt/live/${CERTBOT_DOMAIN}/cert.pem" ${CERT_DIR}/cert.pem
  cp "/etc/letsencrypt/live/${CERTBOT_DOMAIN}/chain.pem" ${CERT_DIR}/chain.pem
  cp "/etc/letsencrypt/live/${CERTBOT_DOMAIN}/fullchain.pem" ${CERT_DIR}/fullchain.pem
  cp "/etc/letsencrypt/live/${CERTBOT_DOMAIN}/privkey.pem" ${CERT_DIR}/privkey.pem

  # Set ownership to Mosquitto
  chown mosquitto: ${CERT_DIR}/cert.pem ${CERT_DIR}/chain.pem ${CERT_DIR}/fullchain.pem ${CERT_DIR}/privkey.pem

  # Ensure permissions are restrictive
  chmod 0600 ${CERT_DIR}/cert.pem ${CERT_DIR}/chain.pem ${CERT_DIR}/fullchain.pem ${CERT_DIR}/privkey.pem

  ls -la ${CERT_DIR}
  ps -a

  mosquitto -c /mosquitto/config/mosquitto.conf
else
  ps -a

  mosquitto -c /mosquitto/config/mosquitto-no-security.conf
fi

sleep infinity