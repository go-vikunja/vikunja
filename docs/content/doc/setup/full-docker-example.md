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

While Vikunja has support to use redis as a caching backend, you'll probably not need it unless you're using Vikunja 
with more than a handful of users.

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

## PostgreSQL

Vikunja supports postgres, mysql and sqlite as a database backend. The examples on this page use mysql with a mariadb container.
To use postgres as a database backend, change the `db` section of the examples to this:

{{< highlight yaml >}}
db:
  image: postgres:13
  environment:
    POSTGRES_PASSWORD: secret
    POSTGRES_USER: vikunja
  volumes:
    - ./db:/var/lib/postgresql/data
  restart: unless-stopped
{{< /highlight >}}

You'll also need to change the `VIKUNJA_DATABASE_TYPE` to `postgres` on the api container declaration.

<div class="notification is-warning">
<b>NOTE:</b> The mariadb container can sometimes take a while to initialize, especially on the first run. 
During this time, the api container will fail to start at all. It will automatically restart every few seconds.
</div>

## Example without any proxy

This example lets you host Vikunja without any reverse proxy in front of it. This is the absolute minimum configuration 
you need to get something up and running. If you want to host Vikunja on one single port instead of two different ones 
or need tls termination, check out one of the other examples.

Not that you need to change the `VIKUNJA_API_URL` environment variable to the ip (the docker host you're running this on) 
is reachable at. Because the browser you'll use to access the Vikunja frontend uses that url to make the requests, it 
has to be able to reach that ip + port from the outside. Putting everything in a private network won't work.

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
    ports:
      - 3456:3456
    volumes:
      - ./files:/app/vikunja/files
    depends_on:
      - db
    restart: unless-stopped
  frontend:
    image: vikunja/frontend
    ports:
      - 80:80
    environment:
      VIKUNJA_API_URL: http://<your-ip-here>:3456/api/v1
    restart: unless-stopped
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
