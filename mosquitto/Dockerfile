# syntax=docker/dockerfile:1

FROM eclipse-mosquitto:2.0-openssl

# install certbot and openssl (all required to enable certbot with standalone mode)
RUN apk update \
    && apk upgrade \
    && apk add --no-cache \
    certbot openssl

WORKDIR /ac
 
COPY run.sh run.sh

# Scripts to renew let's encrypt certs and to upgrade certbot automatically.
# Filenames in `/etc/periodic` must follow specific rules: https://askubuntu.com/a/611430
# For example you cannot use "dot" and "dash" in filenames, so you must remove file extensions
COPY certbotrenew.sh /etc/periodic/daily/certbotrenew
# you have to make all files in /etc/periodic executables,
# otherwise crond will skip non-exetubale files without throwing errors
RUN chmod +x /etc/periodic/daily/certbotrenew

COPY mosquitto.conf /mosquitto/config/mosquitto.conf
COPY mosquitto-no-security.conf /mosquitto/config/mosquitto-no-security.conf

ENTRYPOINT ["sh", "run.sh"]