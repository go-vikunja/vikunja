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
```

This defines two services, each with their own container:

* A Vikunja service which runs the vikunja api and hosts its frontend.
* A database container which will store all projects, tasks, etc. We're using mariadb here, but you're free to use mysql or postgres if you want.

If you already have a proxy on your host, you may want to check out the [reverse proxy examples]({{< ref "reverse-proxies.md" >}}) to use that.
By default, Vikunja will be exposed on port 3456 on the host.

To change to something different, you'll need to change the `ports` section in the service definition.
The number before the colon is the host port - This is where you can reach vikunja from the outside once all is up and running.

You'll need to change the value of the `VIKUNJA_SERVICE_PUBLICURL` environment variable to the public port or hostname where Vikunja is reachable.

## Ensure adequate file permissions

Vikunja runs as user `1000` and no group by default.

To be able to upload task attachments or change the background of project, Vikunja must be able to write into the `files` directory.
To do this, create the folder and chown it before starting the stack:

```
mkdir $PWD/files
chown 1000 $PWD/files
```

## Run it

Run `sudo docker-compose up` in your directory and take a look at the output you get.
When first started, Vikunja will set up the database and run all migrations etc.
Once it is ready, you should see a message like this one in your console:

```
vikunja_1       | 2024-02-09T14:44:06.990677157+01:00: INFO       ▶ cmd/func29 05d Vikunja version 0.23.0
vikunja_1       | ⇨ http server started on [::]:3456
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

```json
{
	"version": "v0.23.0",
	"frontend_url": "https://try.vikunja.io/",
	"motd": "",
	"link_sharing_enabled": true,
	"max_file_size": "20MB",
	"registration_enabled": true,
	"available_migrators": [
		"vikunja-file",
		"ticktick",
		"todoist"
	],
	"task_attachments_enabled": true,
	"enabled_background_providers": [
		"upload",
		"unsplash"
	],
	"totp_enabled": false,
	"legal": {
		"imprint_url": "",
		"privacy_policy_url": ""
	},
	"caldav_enabled": true,
	"auth": {
		"local": {
			"enabled": true
		},
		"openid_connect": {
			"enabled": false,
			"providers": null
		}
	},
	"email_reminders_enabled": true,
	"user_deletion_enabled": true,
	"task_comments_enabled": true,
	"demo_mode_enabled": true,
	"webhooks_enabled": true
}
```

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
