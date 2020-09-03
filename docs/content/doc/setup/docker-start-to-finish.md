---
date: "2020-05-24:00:00+02:00"
title: "Docker Walkthrough"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Setup with docker from start to finish

This tutorial assumes basic knowledge of docker.
It is aimed at beginners and should get you up and running quickly.

We'll use [docker compose](https://docs.docker.com/compose/) to make handling the bunch of containers easier.

> If you have any issues setting up vikunja, please don't hesitate to reach out to us via [matrix](https://riot.im/app/#/room/!dCRiCiLaCCFVNlDnYs:matrix.org?via=matrix.org), the [community forum](https://community.vikunja.io/) or even [email](mailto:hello@vikunja.io).

{{< table_of_contents >}}

## Preparations (optional)

Create a directory for the project where all data and the compose file will live in.

## Create all necessary files

Create a `docker-compose.yml` file with the following contents in your directory:

{{< highlight yaml >}}
version: '3'

services:
  db:
    image: mariadb:10
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    environment:
      MYSQL_ROOT_PASSWORD: supersecret
      MYSQL_DATABASE: vikunja
    volumes:
      - ./db:/var/lib/mysql
    restart: unless-stopped
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

This defines four services, each with their own container:

* An api service which runs the vikunja api. Most of the core logic lives here.
* The frontend which will make vikunja actually usable for most people.
* A database container which will store all lists, tasks, etc. We're using mariadb here, but you're free to use mysql or postgres if you want.
* A proxy service which makes the frontend and api available on the same port, redirecting all requests to `/api` to the api container. 
If you already have a proxy on your host, you may want to check out the [reverse proxy examples]() to use that.
By default, it uses port 80 on the host.
To change to something different, you'll need to change the `ports` section in the service definition.
The number before the colon is the host port - This is where you can reach vikunja from the outside once all is up and running.

For the proxy service we'll need another bit of configuration.
Create an `nginx.conf` in your directory (next to the `docker-compose.yml` file) and add the following contents to it:

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

This is a simple proxy configuration which will forward all requests to `/api/` to the api container and everything else to the frontend.

<div class="notification is-info">
<b>NOTE:</b> Even if you want to make your installation available under a different port, you don't need to change anything in this configuration.
</div>

<div class="notification is-warning">
<b>NOTE:</b> If you change the max upload size in Vikunja's settings, you'll need to also change the <code>client_max_body_size</code> in the nginx proxy config.
</div>

## Run it

Run `sudo docker-compose up` in your directory and take a look at the output you get.
When first started, Vikunja will set up the database and run all migrations etc.
Once it is ready, you should see a message like this one in your console:

```
api_1       | 2020-05-24T11:15:37.560386009Z: INFO	▶ cmd/func1 025 Vikunja version 0.13.1+19-e9bc3246ce, built at Sun, 24 May 2020 11:10:36 +0000
api_1       | ⇨ http server started on [::]:3456
```

This indicates all setup has been successful.
If you get any errors, see below:

### Troubleshooting

Vikunja might not run on the first try.
There are a few potential issues that could be causing this.

#### No connection to the database

Indicated by an error message like this one from the api container:

```
2020/05/23 15:37:59 Config File "config" Not Found in "[/app/vikunja /etc/vikunja /app/vikunja/.config/vikunja]"
2020/05/23 15:37:59 Using default config.
2020-05-23T15:37:59.974435725Z: CRITICAL	▶ migration/Migrate 002 Migration failed: dial tcp 172.19.0.2:3306: connect: connection refused
```

Especially when using mysql, this can happen on first start, because the mysql database container will take a few seconds to start.
Vikunja does not know the container is not ready, therefore it will just try to connect to the db, fail since it is not ready and exit.

If you're using the docker compose example from above, you may notice the `restart: unless-stopped` option at the api service.
This tells docker to restart the api container if it exits, unless you explicitly stop it.
Therefore, it should "magically fix itself" by automatically restarting the container.

After a few seconds (or minutes) you should see a log message like this one from the mariadb container:

```
2020-05-24 11:42:15 0 [Note] mysqld: ready for connections.
Version: '10.4.12-MariaDB-1:10.4.12+maria~bionic'  socket: '/var/run/mysqld/mysqld.sock'  port: 3306  mariadb.org binary distribution
```

The next restart of Vikunja should be successful.
If not, there might be a different error or a bug with Vikunja, please reach out to us in that case.

(If you have an idea about how we could improve this, we'd like to hear it!)

#### "Not a directory"

If you get an error like this one:

```
ERROR: for vikunja_proxy_1 Cannot start service proxy: OCI runtime create failed: container_linux.go:349: starting container process caused "process_linux.go:449: container init caused \"rootfs_linux.go:58: mounting \\\"vikunja/nginx.conf\\\" to rootfs \\\"/var/lib/docker/overlay2/9c8b8f9419c29dad0d1233fbb0a3c36cf403dabd7a55d6f0a47b0c1dd6029994/merged\\\" at \\\"/var/lib/docker/overlay2/9c8b8f9419c29dad0d1233fbb0a3c36cf403dabd7a55d6f0a47b0c1dd6029994/merged/etc/nginx/conf.d/default.conf\\\" caused \\\"not a directory\\\"\"": unknown: Are you trying to mount a directory onto a file (or vice-versa)? Check if the specified host path exists and is the expected type
```

this means docker tried to mount a directory from the host to a file in the container.
This can happen if you did not create the `nginx.conf` file.
Because there is a volume mount for it in the `docker-compose.yml`, Docker will create a folder because non exists, assuming you want to mount a folder into the container.

To fix this, create the file and restart the containers again.

#### Migration failed: commands out of sync

If you get an error like this one:

```
2020/05/23 15:53:38 Config File "config" Not Found in "[/app/vikunja /etc/vikunja /app/vikunja/.config/vikunja]"
2020/05/23 15:53:38 Using default config.
2020-05-23T15:53:38.762747276Z: CRITICAL	▶ migration/Migrate 002 Migration failed: commands out of sync. Did you run multiple statements at once?
```

This is a mysql issue.
Currently, we don't have a better solution than to completely wipe the database files and start over.
To do this, first stop everything by running `sudo docker-compose down`, then remove the `db/` folder in your current folder with `sudo rm -rf db` and start the whole stack again with `sudo docker-compose up -d`.

## Try it

Head over to `http://<host-ip or url>/api/v1/info` in a browser.
You should see something like this:

{{< highlight json >}}
{
  "version": "0.13.1+19-e9bc3246ce",
  "frontend_url": "http://localhost:8080/",
  "motd": "test",
  "link_sharing_enabled": true,
  "max_file_size": "20MB",
  "registration_enabled": true,
  "available_migrators": [
    "wunderlist",
    "todoist"
  ],
  "task_attachments_enabled": true
}
{{< /highlight >}}

This shows you can reach the api through the api proxy.

Now head over to `http://<host-ip or url>/` which should show the login mask.

## Make it persistent

Currently, Vikunja runs in foreground in your terminal.
For a real-world scenario this is not the best way.

Back in your terminal, stop the stack by pressing `CTRL-C` on your keyboard.
Then run `sudo docker-compose up -d` in your again.
The `-d` flag at the end of the command will tell docker to run the containers in the background.
If you need to check the logs after that, you can run `sudo docker-compose logs`.

Vikunja does not have any default users, you'll need to register and account.
After that, you can use it.

## Tear it all down

If you want to completely stop all containers run `sudo docker-compose down` in your terminal.

## Improve this guide

We'll happily accept suggestions and improvements for this guide.
Please [reach out to us](https://vikunja.io/contact/) if you have any.
