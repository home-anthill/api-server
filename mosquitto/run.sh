#!/bin/sh
set -e

# if not defined, apply default production server
CERTBOT_SERVER=${CERTBOT_SERVER:?"https://acme-v02.api.letsencrypt.org/directory"}

# read env variables
echo "Env variables:"
echo "CERTBOT_EMAIL = ${CERTBOT_EMAIL}"
echo "CERTBOT_DOMAIN = ${CERTBOT_DOMAIN}"
echo "CERTBOT_SERVER = ${CERTBOT_SERVER}"


# https://techjogging.com/create-letsencrypt-certificate-alpine-nginx.html
echo "Checking 'crond' existence"
rc-service --list | grep -i crond
# rc-service crond start && rc-update add crond

echo "Creating /etc/periodic/weekly/renew-script.sh"
cat << EOF | tee -a /etc/periodic/daily/renew-script.sh > /dev/null
#!/bin/sh
certbot renew
EOF
chmod a+x /etc/periodic/weekly/renew-script.sh
#run-parts --test /etc/periodic/weekly

echo "Creating /etc/periodic/monthly/upgrade-certbot.sh"
cat << EOF | tee -a /etc/periodic/monthly/upgrade-certbot.sh > /dev/null
#!/bin/sh
apk update certbot certbot-nginx openssl
EOF
chmod a+x /etc/periodic/monthly/upgrade-certbot.sh
#run-parts --test /etc/periodic/monthly


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
cp "/etc/letsencrypt/live/${CERTBOT_DOMAIN}/privkey.pem" ${CERT_DIR}/privkey.pem
cp "/etc/letsencrypt/live/${CERTBOT_DOMAIN}/chain.pem" ${CERT_DIR}/chain.pem

# Set ownership to Mosquitto
chown mosquitto: ${CERT_DIR}/cert.pem ${CERT_DIR}/privkey.pem ${CERT_DIR}/chain.pem

# Ensure permissions are restrictive
chmod 0600 ${CERT_DIR}/cert.pem ${CERT_DIR}/privkey.pem ${CERT_DIR}/chain.pem

ls -la ${CERT_DIR}
ps -a

mosquitto -c /mosquitto/config/mosquitto.conf

sleep infinity