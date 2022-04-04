#!/bin/sh

set -e

mkdir certs
cp /etc/letsencrypt/archive/ac-ks89.eu/cert1.pem /certs/cert.pem
cp /etc/letsencrypt/archive/ac-ks89.eu/fullchain1.pem /certs/fullchain.pem
cp /etc/letsencrypt/archive/ac-ks89.eu/privkey1.pem /certs/privkey.pem

chgrp mosquitto /certs/cert.pem
chgrp mosquitto /certs/fullchain.pem
chgrp mosquitto /certs/privkey.pem
chmod 444 /certs/cert.pem
chmod 444 /certs/fullchain.pem
chmod 444 /certs/privkey.pem

ls -la /certs

sleep 2

mosquitto -v -c /mosquitto/config/mosquitto.conf

sleep infinity