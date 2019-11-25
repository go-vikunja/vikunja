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
It uses an nginx container or traefik on the host to proxy backend and frontend into a single port.

For all available configuration options, see [configuration]({{< ref "config.md">}}).

## Example with traefik

This example assumes [traefik](https://traefik.io) in version 1 installed and configured to [use docker as a configuration provider](https://docs.traefik.io/v1.7/configuration/backends/docker/).

### Without redis

{{< highlight yaml >}}
version: '3'

services:
  api:
    image: vikunja/api
    environment:
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: supersecret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_DATABASE_DATABASE: vikunja
    volumes: 
      - ./files:/app/vikunja/files
    networks:
      - web
      - default
    depends_on:
      - db
    restart: unless-stopped
    labels:
      - "traefik.docker.network=web"
      - "traefik.enable=true"
      - "traefik.frontend.rule=Host:vikunja.example.com;PathPrefix:/api/v1"
      - "traefik.port=3456"
      - "traefik.protocol=http"
  frontend:
    image: vikunja/frontend
    labels:
      - "traefik.docker.network=web"
      - "traefik.enable=true"
      - "traefik.frontend.rule=Host:vikunja.example.com;PathPrefix:/"
      - "traefik.port=80"
      - "traefik.protocol=http"
    networks:
      - web
      - default
    restart: unless-stopped
  db:
    image: mariadb:10
    environment:
      MYSQL_ROOT_PASSWORD: supersupersecret 
      MYSQL_USER: vikunja
      MYSQL_PASSWORD: supersecret
      MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
    restart: unless-stopped
    command: --max-connections=1000

networks:
  web:
    external: true
{{< /highlight >}}

### Without redis

{{< highlight yaml >}}
version: '3'

services:
  api:
    image: vikunja/api
    environment:
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: supersecret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_DATABASE_DATABASE: vikunja
      VIKUNJA_REDIS_ENABLED: 1
      VIKUNJA_REDIS_HOST: 'redis:6379'
      VIKUNJA_CACHE_ENABLED: 1
      VIKUNJA_CACHE_TYPE: redis
    volumes: 
      - ./files:/app/vikunja/files
    networks:
      - web
      - default
    depends_on:
      - db
      - redis
    restart: unless-stopped
    labels:
      - "traefik.docker.network=web"
      - "traefik.enable=true"
      - "traefik.frontend.rule=Host:vikunja.example.com;PathPrefix:/api/v1"
      - "traefik.port=3456"
      - "traefik.protocol=http"
  frontend:
    image: vikunja/frontend
    labels:
      - "traefik.docker.network=web"
      - "traefik.enable=true"
      - "traefik.frontend.rule=Host:vikunja.example.com;PathPrefix:/"
      - "traefik.port=80"
      - "traefik.protocol=http"
    networks:
      - web
      - default
    restart: unless-stopped
  db:
    image: mariadb:10
    environment:
      MYSQL_ROOT_PASSWORD: supersupersecret 
      MYSQL_USER: vikunja
      MYSQL_PASSWORD: supersecret
      MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
    restart: unless-stopped
    command: --max-connections=1000
  redis:
    image: redis

networks:
  web:
    external: true
{{< /highlight >}}

## Example with nginx as proxy

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
    volumes: 
      - ./files:/app/vikunja/files
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
    volumes: 
      - ./files:/app/vikunja/files
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