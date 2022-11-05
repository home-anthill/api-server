#!/bin/sh
set -e

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