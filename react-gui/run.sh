#!/bin/sh
set -e

# Let's encrypt certificates
#Certificate is saved at: /etc/letsencrypt/live/<DOMAIN>/fullchain.pem
#Key is saved at:         /etc/letsencrypt/live/<DOMAIN>/privkey.pem

# Default nginx.conf location loaded by nginx
nginx_conf_file=/etc/nginx/nginx.conf

# if not defined, apply default production server
CERTBOT_SERVER=${CERTBOT_SERVER:?"https://acme-v02.api.letsencrypt.org/directory"}

# read env variables
echo "Env variables:"
echo "CERTBOT_EMAIL = ${CERTBOT_EMAIL}"
echo "CERTBOT_DOMAIN = ${CERTBOT_DOMAIN}"
echo "CERTBOT_SERVER = ${CERTBOT_SERVER}"
echo "BASE_NGINX_CONF_FILEPATH = ${BASE_NGINX_CONF_FILEPATH}"
echo "NGINX_CONF_HOSTPATH = ${NGINX_CONF_HOSTPATH}"

echo "Checking 'crond' existence"
rc-service --list | grep -i crond
# rc-service crond start && rc-update add crond

echo "Creating /etc/periodic/weekly/renew-script.sh"
cat << EOF | tee -a /etc/periodic/weekly/renew-script.sh > /dev/null
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


if [ -f "${NGINX_CONF_HOSTPATH}" ]
then
  echo "nginx.conf already in persistent volume '/home/nginx-conf/'"
  echo "Copying nginx.conf from persistent volume to the final destination"
  cp "${NGINX_CONF_HOSTPATH}" "${nginx_conf_file}"
  # certbot renew --dry-run
  # Check here to setup chrontab: https://eff-certbot.readthedocs.io/en/stable/using.html#setting-up-automated-renewal
  # cronjob
  # 0 12 * * * /usr/bin/certbot renew --quiet

  # start nginx with default config file '${nginx_conf_file}'
  # and wait thanks to "daemon off;".
  # Taken from official docker image https://github.com/nginxinc/docker-nginx/blob/92973a30900b2ed881d208d10cadade34bbbab33/Dockerfile-alpine.template#L123
  nginx -g "daemon off;"
else
  echo "nginx.conf not available in persistent volume '/home/nginx-conf/'"
  echo "Copying basic nginx.conf without SSL to the final destination"
  cp "${BASE_NGINX_CONF_FILEPATH}" "${nginx_conf_file}"

  echo "Calling certbot to get certificates and update ${nginx_conf_file} with SSL configuration"
  # --- get let's encrypt certificate from production server (https://letsencrypt.org/docs/rate-limits/)
  certbot --nginx -m "${CERTBOT_EMAIL}" --agree-tos -d ${CERTBOT_DOMAIN},www.${CERTBOT_DOMAIN} -n --server "${CERTBOT_SERVER}"

  # --- get let's encrypt certificate from staging server (https://letsencrypt.org/docs/staging-environment/)
  # certbot --nginx -m "${CERTBOT_EMAIL}" --agree-tos -d ${CERTBOT_DOMAIN},www.${CERTBOT_DOMAIN} -n --server https://acme-staging-v02.api.letsencrypt.org/directory

  echo "Copying nginx.conf updated by certbot to the persistent volume"
  cp "${nginx_conf_file}" "${NGINX_CONF_HOSTPATH}"

  # IMPORTANT
  # certbot already starts nginx automatically, so I quit and restart nginx to be sure that everything will be ok
  nginx -s quit
  sleep 2
  nginx -g "daemon off;"

  # send signal to nginx (supported values: stop, quit, reopen, reload)
  # nginx -s reload;
fi