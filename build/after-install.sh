#!/bin/bash

systemctl enable vikunja.service

# Fix the config to contain proper values
NEW_SECRET=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
sed -i "s/<jwt-secret>/$NEW_SECRET/g" /etc/vikunja/config.yml
sed -i "s/<rootpath>/\/opt\/vikunja\//g" /etc/vikunja/config.yml
sed -i "s/path: \"\.\/vikunja.db\"/path: \"\\/opt\/vikunja\/vikunja.db\"/g" /etc/vikunja/config.yml
