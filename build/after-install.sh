#!/bin/bash

systemctl enable vikunja.service

# Fix the config to contain proper values
NEW_SECRET=$(head -c 512 /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32)
sed -i "s/<jwt-secret>/$NEW_SECRET/g" /etc/vikunja/config.yml
sed -i "s/<rootpath>/\/opt\/vikunja\//g" /etc/vikunja/config.yml
sed -i "s/path: \"\.\/vikunja.db\"/path: \"\\/opt\/vikunja\/vikunja.db\"/g" /etc/vikunja/config.yml
