#!/bin/sh

mkdir certs
cp /etc/letsencrypt/archive/ac-ks89.eu/cert1.pem /certs/cert.pem
cp /etc/letsencrypt/archive/ac-ks89.eu/fullchain1.pem /certs/fullchain.pem
cp /etc/letsencrypt/archive/ac-ks89.eu/privkey1.pem /certs/privkey.pem

chgrp mosquitto /certs/cert.pem
chgrp mosquitto /certs/fullchain.pem
chgrp mosquitto /certs/privkey.pem
chmod g+r /certs/cert.pem
chmod g+r /certs/fullchain.pem
chmod g+r /certs/privkey.pem

ls -la /certs

mosquitto -c /mosquitto/config/mosquitto.conf