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

<div class="notification is-warning">
<b>NOTE:</b> If you intend to run Vikunja with mysql and/or to use non-latin characters 
<a href="{{< ref "utf-8.md">}}">make sure your db is utf-8 compatible</a>.<br/>
All examples on this page already reflect this and do not require additional work.
</div>

{{< table_of_contents >}}

## Redis

To use redis, you'll need to add this to the config examples below:

{{< highlight yaml >}}
version: '3'

services:
  api:
    image: vikunja/api
    environment:
      VIKUNJA_REDIS_ENABLED: 1
      VIKUNJA_REDIS_HOST: 'redis:6379'
      VIKUNJA_CACHE_ENABLED: 1
      VIKUNJA_CACHE_TYPE: redis
    volumes:
      - ./files:/app/vikunja/files
  redis:
    image: redis
{{< /highlight >}}

## Example with traefik 2

This example assumes [traefik](https://traefik.io) version 2 installed and configured to [use docker as a configuration provider](https://docs.traefik.io/providers/docker/).

We also make a few assumtions here which you'll most likely need to adjust for your traefik setup:

* Your domain is `vikunja.example.com`
* The entrypoint you want to make vikunja available from is called `https`
* The tls cert resolver is called `acme`

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
      - "traefik.enable=true"
      - "traefik.http.routers.vikunja-api.rule=Host(`vikunja.example.com`) && PathPrefix(`/api/v1`, `/dav/`, `/.well-known/`)"
      - "traefik.http.routers.vikunja-api.entrypoints=https"
      - "traefik.http.routers.vikunja-api.tls.certResolver=acme"
  frontend:
    image: vikunja/frontend
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.vikunja-frontend.rule=Host(`vikunja.example.com`)"
      - "traefik.http.routers.vikunja-frontend.entrypoints=https"
      - "traefik.http.routers.vikunja-frontend.tls.certResolver=acme"
    networks:
      - web
      - default
    restart: unless-stopped
  db:
    image: mariadb:10
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
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

## Example with traefik 1

This example assumes [traefik](https://traefik.io) in version 1 installed and configured to [use docker as a configuration provider](https://docs.traefik.io/v1.7/configuration/backends/docker/).

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
      - "traefik.frontend.rule=Host:vikunja.example.com;PathPrefix:/api/v1,/dav/,/.well-known"
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
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
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

## Example with nginx as proxy

You'll need to save this nginx configuration on your host under `nginx.conf` 
(or elsewhere, but then you'd need to adjust the proxy mount at the bottom of the compose file):

{{< highlight conf >}}
server {
    listen 80;

    location / {
        proxy_pass http://frontend:80;
    }

    location ~* ^/(api|dav|\.well-known)/ {
        proxy_pass http://api:3456;
        client_max_body_size 20M;
    }
}
{{< /highlight >}}

<div class="notification is-warning">
<b>NOTE:</b> If you change the max upload size in Vikunja's settings, you'll need to also change the <code>client_max_body_size</code> in the nginx proxy config.
</div>

`docker-compose.yml` config:

{{< highlight yaml >}}
version: '3'

services:
  db:
    image: mariadb:10
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    environment:
      MYSQL_ROOT_PASSWORD: supersecret
	  MYSQL_USER: vikunja
	  MYSQL_PASSWORD: secret
	  MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
    restart: unless-stopped
  api:
    image: vikunja/api
    environment:
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: secret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_DATABASE_DATABASE: vikunja
    volumes: 
      - ./files:/app/vikunja/files
    depends_on:
      - db
    restart: unless-stopped
  frontend:
    image: vikunja/frontend
    restart: unless-stopped
  proxy:
    image: nginx
    ports:
      - 80:80
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      - api
      - frontend
    restart: unless-stopped
{{< /highlight >}}

## Example with Caddy v2 as proxy

You will need the following `Caddyfile` on your host (or elsewhere, but then you'd need to adjust the proxy mount at the bottom of the compose file):

{{< highlight conf >}}
vikunja.example.com {
    reverse_proxy /api/* api:3456
    reverse_proxy /.well-known/* api:3456
    reverse_proxy /dav/* api:3456
    reverse_proxy frontend:80
}
{{< /highlight >}}

`docker-compose.yml` config:

{{< highlight yaml >}}
version: '3'

services:
  db:
    image: mariadb:10
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    environment:
      MYSQL_ROOT_PASSWORD: supersecret
	  MYSQL_USER: vikunja
	  MYSQL_PASSWORD: secret
	  MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
    restart: unless-stopped
  api:
    image: vikunja/api
    environment:
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: secret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_DATABASE_DATABASE: vikunja
    volumes: 
      - ./files:/app/vikunja/files
    depends_on:
      - db
    restart: unless-stopped
  frontend:
    image: vikunja/frontend
    restart: unless-stopped
  caddy:
    image: caddy
    restart: unless-stopped
    ports:
        - "80:80"
        - "443:443"
    depends_on:
      - api
      - frontend
    volumes:
        - ./Caddyfile:/etc/caddy/Caddyfile:ro
{{< /highlight >}}
