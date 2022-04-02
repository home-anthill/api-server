#!/bin/sh

#Certificate is saved at: /etc/letsencrypt/live/<DOMAIN>/fullchain.pem
#Key is saved at:         /etc/letsencrypt/live/<DOMAIN>/privkey.pem

if [[ -f /home/nginx-conf/nginx.conf ]]
then
  echo "nginx.conf already in persistent volume '/home/nginx-conf/'"
  echo "Copying nginx.conf from persistent volume to the final destination"
  cp /home/nginx-conf/nginx.conf /etc/nginx/nginx.conf
  # certbot renew --dry-run
  # Check here to setup chrontab: https://eff-certbot.readthedocs.io/en/stable/using.html#setting-up-automated-renewal
  # cronjob
  # 0 12 * * * /usr/bin/certbot renew --quiet

  # start nginx with default config file '/etc/nginx/nginx.conf'
  # and wait thanks to "daemon off;".
  # Taken from official docker image https://github.com/nginxinc/docker-nginx/blob/92973a30900b2ed881d208d10cadade34bbbab33/Dockerfile-alpine.template#L123
  nginx -g "daemon off;"
else
  echo "nginx.conf not available in persistent volume '/home/nginx-conf/'"

  cp /home/nginx.conf /etc/nginx/nginx.conf

  # --- get let's encrypt certificate from production server (https://letsencrypt.org/docs/rate-limits/)
  certbot --nginx -m stefano.cappa.ks89@gmail.com --agree-tos -d ac-ks89.eu,www.ac-ks89.eu -n --server https://acme-v02.api.letsencrypt.org/directory

  # --- get let's encrypt certificate from staging server (https://letsencrypt.org/docs/staging-environment/)
  # certbot --nginx -m stefano.cappa.ks89@gmail.com --agree-tos -d ac-ks89.eu,www.ac-ks89.eu -n --server https://acme-staging-v02.api.letsencrypt.org/directory

  echo "Copying nginx.conf from persistent volume to the final destination"
  cp /etc/nginx/nginx.conf /home/nginx-conf/nginx.conf

  # IMPORTANT
  # certbot already starts nginx automatically, so I quit and restart nginx to be sure that everything will be ok
  nginx -s quit
  sleep 2
  nginx -g "daemon off;"

  # send signal to nginx (supported values: stop, quit, reopen, reload)
  # nginx -s reload;
fi