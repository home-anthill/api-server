#!/bin/sh
set -e

# Default nginx.conf location loaded by nginx
nginx_conf_file=/etc/nginx/nginx.conf

# read env variables
echo "Env variables:"
echo "BASE_NGINX_CONF_FILEPATH = ${BASE_NGINX_CONF_FILEPATH}"
echo "NGINX_CONF_HOSTPATH = ${NGINX_CONF_HOSTPATH}"

if [ ${CERTBOT_EMAIL+x} ] && [ ${CERTBOT_DOMAIN+x} ] && [ ${CERTBOT_SERVER+x} ]; then
   # if not defined, apply default production server
  CERTBOT_SERVER=${CERTBOT_SERVER:?"https://acme-v02.api.letsencrypt.org/directory"}

  echo "Printing 'crontab'"
  crontab -l
  echo "Run crond as background process"
  crond -b -l 0 -L /ac/crond.log

  echo "CERTBOT_EMAIL = ${CERTBOT_EMAIL}"
  echo "CERTBOT_DOMAIN = ${CERTBOT_DOMAIN}"
  echo "CERTBOT_SERVER = ${CERTBOT_SERVER}"

  echo "Starting nginx with SSL"
  # Let's encrypt certificates
  #Certificate is saved at: /etc/letsencrypt/live/<DOMAIN>/fullchain.pem
  #Key is saved at:         /etc/letsencrypt/live/<DOMAIN>/privkey.pem

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
else
  echo "Starting nginx without SSL"
  if [ -f "${NGINX_CONF_HOSTPATH}" ]; then
      echo "nginx.conf already in persistent volume '/home/nginx-conf/'"
      echo "Copying nginx.conf from persistent volume to the final destination"
      cp "${NGINX_CONF_HOSTPATH}" "${nginx_conf_file}"

      # start nginx with default config file '${nginx_conf_file}'
      # and wait thanks to "daemon off;".
      # Taken from official docker image https://github.com/nginxinc/docker-nginx/blob/92973a30900b2ed881d208d10cadade34bbbab33/Dockerfile-alpine.template#L123
      nginx -g "daemon off;"
  else
    echo "nginx.conf not available in persistent volume '/home/nginx-conf/'"
    echo "Copying basic nginx.conf without SSL to the final destination"
    cp "${BASE_NGINX_CONF_FILEPATH}" "${nginx_conf_file}"

    echo "Copying nginx.conf updated by certbot to the persistent volume"
    cp "${nginx_conf_file}" "${NGINX_CONF_HOSTPATH}"

    nginx -s quit
    sleep 2
    nginx -g "daemon off;"
  fi
fi