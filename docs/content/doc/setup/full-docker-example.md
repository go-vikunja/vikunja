---
date: "2019-02-12:00:00+02:00"
title: "Full docker example"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Full docker example

This docker compose configuration will run Vikunja with backend and frontend with a mariadb as database.
It uses an nginx container to proxy backend and frontend into a single port.

You'll need to save this nginx configuration on your host under `nginx.conf` 
(or elsewhere, but then you'd need to adjust the proxy mount at the bottom of the compose file):

{{< highlight conf >}}
server {
    listen 80;

    location / {
        proxy_pass http://frontend:80;
    }

    location /api/ {
        proxy_pass http://api:3456;
    }
}
{{< /highlight >}}

### Without redis

{{< highlight yaml >}}
version: '3'

services:
  db:
    image: mariadb:10
    environment:
      MYSQL_ROOT_PASSWORD: supersecret
      MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
  api:
    image: vikunja/api
    environment:
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: supersecret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: root
      VIKUNJA_DATABASE_DATABASE: vikunja
    depends_on:
      - db
  frontend:
    image: vikunja/frontend
  proxy:
    image: nginx
    ports:
      - 80:80
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      - api
      - frontend
{{< /highlight >}}

### With redis

{{< highlight yaml >}}
version: '3'

services:
  db:
    image: mariadb:10
    environment:
      MYSQL_ROOT_PASSWORD: supersecret
      MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
  redis:
    image: redis
  api:
    image: vikunja/api
    environment:
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: supersecret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: root
      VIKUNJA_DATABASE_DATABASE: vikunja
      VIKUNJA_REDIS_ENABLED: 1
      VIKUNJA_REDIS_HOST: 'redis:6379'
      VIKUNJA_CACHE_ENABLED: 1
      VIKUNJA_CACHE_TYPE: redis
    depends_on:
      - db
      - redis
  frontend:
    image: vikunja/frontend
  proxy:
    image: nginx
    ports:
      - 80:80
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      - api
      - frontend
{{< /highlight >}}