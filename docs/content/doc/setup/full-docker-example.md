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

This docker compose configuration will run Vikunja with a mariadb database.
It uses a proxy configuration to make it available under a domain.

For all available configuration options, see [configuration]({{< ref "config.md">}}).

After registering all your users, you might also want to [disable the user registration]({{<ref "config.md">}}#enableregistration).

<div class="notification is-warning">
<b>NOTE:</b> If you intend to run Vikunja with mysql and/or to use non-latin characters
<a href="{{< ref "utf-8.md">}}">make sure your db is utf-8 compatible</a>.<br/>
All examples on this page already reflect this and do not require additional work.
</div>

{{< table_of_contents >}}

## File permissions

Vikunja runs as user `1000` and no group by default.
You can use Docker's [`--user`](https://docs.docker.com/engine/reference/run/#user) flag to change that.

You must ensure Vikunja is able to write into the `files` directory.
To do this, create the folder and chown it before starting the stack:

```
mkdir $PWD/files
chown 1000 $PWD/files
```

You'll need to do this before running any of the examples on this page.

Vikunja will not try to acquire ownership of the files folder, as that would mean it had to run as root.

## PostgreSQL

Vikunja supports postgres, mysql and sqlite as a database backend. The examples on this page use mysql with a mariadb container.
To use postgres as a database backend, change the `db` section of the examples to this:

```yaml
db:
  image: postgres:16
  environment:
    POSTGRES_PASSWORD: changeme
    POSTGRES_USER: vikunja
  volumes:
    - ./db:/var/lib/postgresql/data
  restart: unless-stopped
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -h localhost -U $$POSTGRES_USER"]
    interval: 2s
```

You'll also need to change the `VIKUNJA_DATABASE_TYPE` to `postgres` on the api container declaration.

## Sqlite

Vikunja supports postgres, mysql and sqlite as a database backend. The examples on this page use mysql with a mariadb container.
To use sqlite as a database backend, change the `api` section of the examples to this:

```yaml
vikunja:
  image: vikunja/vikunja
  environment:
    VIKUNJA_SERVICE_JWTSECRET: <a super secure random secret>
    VIKUNJA_SERVICE_PUBLICURL: http://<your public frontend url with slash>/
    # Note the default path is /app/vikunja/vikunja.db.
    # This config variable moves it to a different folder so you can use a volume and 
    # store the database file outside the container so state is persisted even if the container is destroyed.
    VIKUNJA_DATABASE_PATH: /db/vikunja.db
  ports:
    - 3456:3456
  volumes:
    - ./files:/app/vikunja/files
    - ./db:/db
  restart: unless-stopped
```

The default path Vikunja uses for sqlite is relative to the binary, which in the docker container would be `/app/vikunja/vikunja.db`.
The `VIKUNJA_DATABASE_PATH` environment variable moves changes it so that the database file is stored in a volume at `/db`, to persist state across restarts.

You'll also need to remove or change the `VIKUNJA_DATABASE_TYPE` to `sqlite` on the container declaration.

You can also remove the db section.

To run the container, you need to create the directories first and make sure they have all required permissions:

```
mkdir $PWD/files $PWD/db
chown 1000 $PWD/files $PWD/db
```

<div class="notification is-warning">
<b>NOTE:</b> If you'll use your instance with more than a handful of users, we recommend using mysql or postgres.
</div>

## Example without any proxy

This example lets you host Vikunja without any reverse proxy in front of it. 
This is the absolute minimum configuration you need to get something up and running. 
If you want to make Vikunja available on a domain or need tls termination, check out one of the other examples.

Note that you need to change the [`VIKUNJA_SERVICE_PUBLICURL`]({{< ref "config.md" >}}#publicurl) environment variable to the public ip or hostname including the port (the docker host you're running this on) is reachable at, prefixed with `http://`.
Because the browser you'll use to access the Vikunja frontend uses that url to make the requests, it has to be able to reach it from the outside. 

<div class="notification is-warning">
<b>NOTE:</b> You must ensure Vikunja has write permissions on the `files` directory before starting the stack.
To do this, <a href="#file-permissions">check out the related commands here</a>.
</div>

```yaml
version: '3'

services:
  vikunja:
    image: vikunja/vikunja
    environment:
      VIKUNJA_SERVICE_PUBLICURL: http://<the public ip or host where vikunja is reachable>
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: changeme
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_DATABASE_DATABASE: vikunja
      VIKUNJA_SERVICE_JWTSECRET: <a super secure random secret>
    ports:
      - 3456:3456
    volumes:
      - ./files:/app/vikunja/files
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped
  db:
    image: mariadb:10
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    environment:
      MYSQL_ROOT_PASSWORD: supersecret
      MYSQL_USER: vikunja
      MYSQL_PASSWORD: changeme
      MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h localhost -u $$MYSQL_USER --password=$$MYSQL_PASSWORD"]
      interval: 2s
      start_period: 30s
```

## Example with Traefik 2

This example assumes [traefik](https://traefik.io) version 2 installed and configured to [use docker as a configuration provider](https://docs.traefik.io/providers/docker/).

We also make a few assumptions here which you'll most likely need to adjust for your traefik setup:

* Your domain is `vikunja.example.com`
* The entrypoint you want to make vikunja available from is called `https`
* The tls cert resolver is called `acme`

<div class="notification is-warning">
<b>NOTE:</b> You must ensure Vikunja has write permissions on the `files` directory before starting the stack.
To do this, <a href="#file-permissions">check out the related commands here</a>.
</div>

```yaml
version: '3'

services:
  vikunja:
    image: vikunja/vikunja
    environment:
      VIKUNJA_SERVICE_PUBLICURL: http://<the public url where vikunja is reachable>
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: changeme
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_DATABASE_DATABASE: vikunja
      VIKUNJA_SERVICE_JWTSECRET: <a super secure random secret>
    volumes: 
      - ./files:/app/vikunja/files
    networks:
      - web
      - default
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped
    labels:
      - "traefik.enable=true"
      - "traefik.docker.network=web"
      - "traefik.http.routers.vikunja.rule=Host(`vikunja.example.com`)"
      - "traefik.http.routers.vikunja.entrypoints=https"
      - "traefik.http.routers.vikunja.tls.certResolver=acme"
  db:
    image: mariadb:10
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    environment:
      MYSQL_ROOT_PASSWORD: supersupersecret 
      MYSQL_USER: vikunja
      MYSQL_PASSWORD: changeme
      MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h localhost -u $$MYSQL_USER --password=$$MYSQL_PASSWORD"]
      interval: 2s
      start_period: 30s

networks:
  web:
    external: true
```

## Example with Caddy v2 as proxy

You will need the following `Caddyfile` on your host (or elsewhere, but then you'd need to adjust the proxy mount at the bottom of the compose file):

```conf
vikunja.example.com {
    reverse_proxy api:3456
}
```

Note that you need to change the [`VIKUNJA_SERVICE_PUBLICURL`]({{< ref "config.md" >}}#publicurl) environment variable to the ip (the docker host you're running this on) is reachable at.
Because the browser you'll use to access the Vikunja frontend uses that url to make the requests, it has to be able to reach that ip + port from the outside.

<div class="notification is-warning">
<b>NOTE:</b> You must ensure Vikunja has write permissions on the `files` directory before starting the stack.
To do this, <a href="#file-permissions">check out the related commands here</a>.
</div>

Docker Compose config:

```yaml
version: '3'

services:
  vikunja:
    image: vikunja/vikunja
    environment:
      VIKUNJA_SERVICE_PUBLICURL: http://<the public url where vikunja is reachable>
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: changeme
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_DATABASE_DATABASE: vikunja
      VIKUNJA_SERVICE_JWTSECRET: <a super secure random secret>
    ports:
      - 3456:3456
    volumes:
      - ./files:/app/vikunja/files
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped
  db:
    image: mariadb:10
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    environment:
      MYSQL_ROOT_PASSWORD: supersecret
      MYSQL_USER: vikunja
      MYSQL_PASSWORD: changeme
      MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h localhost -u $$MYSQL_USER --password=$$MYSQL_PASSWORD"]
      interval: 2s
      start_period: 30s
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
```

## Setup on a Synology NAS

There is a proxy preinstalled in DSM, so if you want to access Vikunja from outside,
you need to prepare a proxy rule the Vikunja Service.

![Synology Proxy Settings](/docs/synology-proxy-1.png)

You should also add 2 empty folders for mariadb and vikunja inside Synology's
docker main folders:

* Docker
  * vikunja
  * mariadb

Synology has its own GUI for managing Docker containers, but it's easier via docker compose.

To do that, you can

* Either activate SSH and paste the adapted compose file in a terminal (using Putty or similar)
* Without activating SSH as a "custom script" (go to Control Panel / Task Scheduler / Create / Scheduled Task / User-defined script)
* Without activating SSH, by using Portainer (you have to install first, check out [this tutorial](https://www.portainer.io/blog/how-to-install-portainer-on-a-synology-nas) for example):
  1. Go to **Dashboard / Stacks** click the button **"Add Stack"**
  2. Give it the name Vikunja and paste the adapted docker compose file
  3. Deploy the Stack with the "Deploy Stack" button:

![Portainer Stack deploy](/docs/synology-proxy-2.png)

The docker-compose file we're going to use is exactly the same from the [example without any proxy](#example-without-any-proxy) above.

You may want to change the volumes to match the rest of your setup.

After registering all your users, you might also want to [disable the user registration]({{<ref "config.md">}}#enableregistration).

<div class="notification is-warning">
<b>NOTE:</b> You must ensure Vikunja has write permissions on the `files` directory before starting the stack.
To do this, <a href="#file-permissions">check out the related commands here</a>.
</div>

## Redis

While Vikunja has support to use redis as a caching backend, you'll probably not need it unless you're using Vikunja with more than a handful of users.

To use redis, you'll need to add this to the config examples below:

```yaml
version: '3'

services:
  vikunja:
    image: vikunja/vikunja
    environment:
      VIKUNJA_REDIS_ENABLED: 1
      VIKUNJA_REDIS_HOST: 'redis:6379'
      VIKUNJA_CACHE_ENABLED: 1
      VIKUNJA_CACHE_TYPE: redis
    volumes:
      - ./files:/app/vikunja/files
  redis:
    image: redis
```
