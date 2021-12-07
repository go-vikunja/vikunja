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

<div class="notification is-warning">
<b>NOTE:</b> If you intend to run Vikunja with mysql and/or to use non-latin characters 
<a href="{{< ref "utf-8.md">}}">make sure your db is utf-8 compatible</a>.
</div>

{{< table_of_contents >}}

## Install from binary

Download a copy of Vikunja from the [download page](https://vikunja.io/en/download/) for your architecture.

{{< highlight bash >}}
wget <download-url>
{{< /highlight >}}

### Verify the GPG signature

Starting with version `0.7`, all releases are signed using pgp.
Releases from `main` will always be signed.

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

Save the following service file to `/etc/systemd/system/vikunja.service` and adapt it to your needs:

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
This will expose vikunja on port `3456` on the host running the container.

You can mount a local configuration like so:

{{< highlight bash >}}
docker run -p 3456:3456 -v /path/to/config/on/host.yml:/app/vikunja/config.yml:ro vikunja/api
{{< /highlight >}}

Though it is recommended to use eviroment variables or `.env` files to configure Vikunja in docker.
See [config]({{< ref "config.md">}}) for a list of available configuration options.

### Files volume

By default the container stores all files uploaded and used through vikunja inside of `/app/vikunja/files` which is created as a docker volume.
You should mount the volume somewhere to the host to permanently store the files and don't loose them if the container restarts.

### Setting user and group id of the user running vikunja

You can set the user and group id of the user running vikunja with the `PUID` and `PGID` evironment variables.
This follows the pattern used by [the linuxserver.io](https://docs.linuxserver.io/general/understanding-puid-and-pgid) docker images.

This is useful to solve general permission problems when host-mounting volumes such as the volume used for task attachments.

### Docker compose

To run the backend with a mariadb database you can use this example [docker-compose](https://docs.docker.com/compose/) file:

{{< highlight yaml >}}
version: '2'
services:
  api:
    image: vikunja/api:latest
    environment:
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: secret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_SERVICE_JWTSECRET: <generated secret>
    volumes:
      - ./files:/app/vikunja/files
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

## FreeBSD / FreeNAS

Unfortunately, we currently can't provide pre-built binaries for FreeBSD.
As a workaround, it is possible to compile vikunja for FreeBSD directly on a FreeBSD machine, a guide is available below:

*Thanks to HungrySkeleton who originally created this guide [in the forum](https://community.vikunja.io/t/freebsd-support/69/11).*

### Jail Setup

1. Create jail named ```vikunja```
2. Set jail properties to 'auto start'
3. Mount storage (```/mnt``` to ```jailData/vikunja```)
4. Start jail & SSH into it

### Installing packages

{{< highlight bash >}}
pkg update && pkg upgrade -y
pkg install nano git go gmake
go install github.com/magefile/mage
{{< /highlight >}}

### Clone vikunja repo

{{< highlight bash >}}
mkdir /mnt/GO/code.vikunja.io
cd /mnt/GO/code.vikunja.io
git clone https://code.vikunja.io/api
cd /mnt/GO/code.vikunja.io/api
{{< /highlight >}}

### Compile binaries

{{< highlight bash >}}
go install
mage build
{{< /highlight >}}

### Create folder to install backend server into

{{< highlight bash >}}
mkdir /mnt/backend
cp /mnt/GO/code.vikunja.io/api/vikunja /mnt/backend/vikunja
cd /mnt/backend
chmod +x /mnt/backend/vikunja
{{< /highlight >}}

### Set vikunja to boot on startup

{{< highlight bash >}}
nano /etc/rc.d/vikunja
{{< /highlight >}}

Then paste into the file:

{{< highlight bash >}}
#!/bin/sh

. /etc/rc.subr

name=vikunja
rcvar=vikunja_enable

command="/mnt/backend/${name}"

load_rc_config $name
run_rc_command "$1"
{{< /highlight >}}

Save and exit.  Then execute:

{{< highlight bash >}}
chmod +x /etc/rc.d/vikunja
nano /etc/rc.conf
{{< /highlight >}}

Then add line to bottom of file:

{{< highlight bash >}}
vikunja_enable="YES"
{{< /highlight >}}

Test vikunja now works with

{{< highlight bash >}}
service vikunja start
{{< /highlight >}}

The API is now available through IP:

```
192.168.1.XXX:3456
```

## Configuration

See [available configuration options]({{< ref "config.md">}}).

## Default Password

After successfully installing Vikunja, there is no default user or password.
You only need to register a new account and set all the details when creating it.
