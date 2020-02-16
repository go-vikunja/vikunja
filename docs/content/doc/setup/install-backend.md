---
date: "2019-02-12:00:00+02:00"
title: "Install Backend"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Backend

## Install from binary

Download a copy of Vikunja from the [download page](https://vikunja.io/en/download/) for your architecture.

{{< highlight bash >}}
wget <download-url>
{{< /highlight >}}

### Verify the GPG signature

Starting with version `0.7`, all releases are signed using pgp.
Releases from `master` will always be signed.

To validate the downloaded zip file use the signiture file `.asc` and the key `FF054DACD908493A`:

{{< highlight bash >}}
gpg --keyserver keyserver.ubuntu.com --recv FF054DACD908493A
gpg --verify vikunja-0.7-linux-amd64-full.zip.asc vikunja-0.7-linux-amd64-full.zip
{{< /highlight >}}

### Set it up

Once you've verified the signature, you need to unzip it and make it executable, you'll also need to 
create a symlink to it so you can execute Vikunja by typing `vikunja` on your system.
We'll install vikunja to `/opt/vikunja`, change the path where needed if you want to install it elsewhere.

{{< highlight bash >}}
mkdir -p /opt/vikunja
unzip <vikunja-zip-file> -d /opt/vikunja
chmod +x /opt/vikunja
ln -s /opt/vikunja/vikunja /usr/bin/vikunja
{{< /highlight >}}

### Systemd service

Take the following `service` file and adapt it to your needs:

{{< highlight service >}}
[Unit]
Description=Vikunja
After=syslog.target
After=network.target
# Depending on how you configured Vikunja, you may want to uncomment these:
#Requires=mysql.service
#Requires=mariadb.service
#Requires=postgresql.service
#Requires=redis.service

[Service]
RestartSec=2s
Type=simple
WorkingDirectory=/opt/vikunja
ExecStart=/usr/bin/vikunja
Restart=always
# If you want to bind Vikunja to a port below 1024 uncomment
# the two values below
###
#CapabilityBoundingSet=CAP_NET_BIND_SERVICE
#AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
{{< /highlight >}}

If you've installed Vikunja to a directory other than `/opt/vikunja`, you need to adapt `WorkingDirectory` accordingly.

Save the file to `/etc/systemd/system/vikunja.service`

After you made all nessecary modifications, it's time to start the service:

{{< highlight bash >}}
sudo systemctl enable vikunja
sudo systemctl start vikunja
{{< /highlight >}}

### Build from source

To build vikunja from source, see [building from source]({{< ref "build-from-source.md">}}).

### Updating

Simply replace the binary and templates with the new version, then restart Vikunja.
It will automatically run all nessecary database migrations.
**Make sure to take a look at the changelog for the new version to not miss any manual steps the update may involve!**

## Docker

(Note: this assumes some familarity with docker)

Usage with docker is pretty straightforward:

{{< highlight bash >}}
docker run -p 3456:3456 vikunja/api
{{< /highlight >}}

to run with a standard configuration.
This will expose 

You can mount a local configuration like so:

{{< highlight bash >}}
docker run -p 3456:3456 -v /path/to/config/on/host.yml:/app/vikunja/config.yml:ro vikunja/api
{{< /highlight >}}

Though it is recommended to use eviroment variables or `.env` files to configure Vikunja in docker.
See [config]({{< ref "config.md">}}) for a list of available configuration options.

### Docker compose

To run the backend with a mariadb database you can use this example [docker-compose](https://docs.docker.com/compose/) file:

{{< highlight yaml >}}
version: '2'
services:
  api:
    image: vikunja/api:latest
    environment:
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: supersecret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: root
      VIKUNJA_SERVICE_JWTSECRET: <generated secret>
  db:
    image: mariadb:10
    environment:
      MYSQL_ROOT_PASSWORD: supersecret
      MYSQL_DATABASE: vikunja
    volumes:
    - ./db:/var/lib/mysql
{{< /highlight >}}

See [full docker example]({{< ref "full-docker-example.md">}}) for more varations of this config.

## Debian packages

Since version 0.7 Vikunja is also released as debian packages.

To install these, grab a copy from [the download page](https://vikunja.io/en/download/) and run

{{< highlight bash >}}
dpkg -i vikunja.deb
{{< /highlight >}}

This will install the backend to `/opt/vikunja`.
To configure it, use the config file in `/etc/vikunja/config.yml`.

## Configuration

See [available configuration options]({{< ref "config.md">}}).
