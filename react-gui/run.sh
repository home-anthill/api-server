#!/bin/sh

#Certificate is saved at: /etc/letsencrypt/live/ac-ks89.eu/fullchain.pem
#Key is saved at:         /etc/letsencrypt/live/ac-ks89.eu/privkey.pem

if [[ -f /home/nginx-conf/nginx.conf ]]
then
  echo "nginx.conf already in persistent volume '/home/nginx-conf/'"
  echo "Copying nginx.conf from persistent volume to the final destination"
  cp /home/nginx-conf/nginx.conf /etc/nginx/nginx.conf
  # certbot renew --dry-run
  # cronjob
  # 0 12 * * * /usr/bin/certbot renew --quiet
else
  echo "nginx.conf not available in persistent volume '/home/nginx-conf/'"

  cp /home/nginx.conf /etc/nginx/nginx.conf
  # certbot --nginx -m stefano.cappa.ks89@gmail.com --agree-tos -d ac-ks89.eu -n --test-cert
  certbot --nginx -m stefano.cappa.ks89@gmail.com --agree-tos -d ac-ks89.eu -d www.ac-ks89.eu -n --server https://acme-staging-v02.api.letsencrypt.org/directory

  echo "Copying nginx.conf from persistent volume to the final destination"
  cp /etc/nginx/nginx.conf /home/nginx-conf/nginx.conf
fi

nginx -s reload;

sleep infinity