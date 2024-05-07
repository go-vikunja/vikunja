---
date: "2019-02-12:00:00+02:00"
title: "Installing"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
    weight: 10
---

# Installing

Architecturally, Vikunja is made up of two parts: [API](https://code.vikunja.io/api) and [frontend](https://code.vikunja.io/api/frontend).

Both are bundled into one single deployable binary (or docker container).
That means you only need to install one thing to be able to use Vikunja.

You can also:

* Use the desktop app, which is essentially the web frontend packaged for easy installation on desktop devices
* Use the mobile app only, but as of right now it only supports the very basic features of Vikunja

<div class="notification is-warning">
<b>NOTE:</b> If you intend to run Vikunja with mysql and/or to use non-latin characters 
<a href="{{< ref "utf-8.md">}}">make sure your db is utf-8 compatible</a>.
</div>

Vikunja can be installed in various ways.
This document provides an overview and instructions for the different methods:

* [Installing from binary (manual)](#install-from-binary)
* [Build from source]({{< ref "build-from-source.md">}})
* [Docker](#docker)
* [Debian](#debian-packages)
* [RPM](#rpm)
* [FreeBSD](#freebsd--freenas)
* [Kubernetes]({{< ref "k8s.md" >}})

And after you installed Vikunja, you may want to check out these other resources:

* [Configuration]({{< ref "config.md">}})
* [UTF-8 Settings]({{< ref "utf-8.md">}})
* [Reverse proxies]({{< ref "reverse-proxies.md">}})
* [Full docker example]({{< ref "full-docker-example.md">}})
* [Backups]({{< ref "backups.md">}})

## Install from binary

Download a copy of Vikunja from the [download page](https://dl.vikunja.io/vikunja) for your architecture.

```
wget <download-url>
```

### Verify the GPG signature

All releases are signed using GPG.

To validate the downloaded zip file use the signiture file `.asc` and the key `FF054DACD908493A`:

```
gpg --keyserver keyserver.ubuntu.com --recv FF054DACD908493A
gpg --verify vikunja-<vikunja version>-linux-amd64-full.zip.asc vikunja-<vikunja version>-linux-amd64-full.zip
```

### Set it up

Once you've verified the signature, you need to unzip and make it executable.
You'll also need to create a symlink to the binary, so that you can execute Vikunja by typing `vikunja` on your system.
We'll install vikunja to `/opt/vikunja`, change the path where needed if you want to install it elsewhere.

Run these commands to install it:

```
mkdir -p /opt/vikunja
unzip <vikunja-zip-file> -d /opt/vikunja
chmod +x /opt/vikunja
sudo ln -s /opt/vikunja/vikunja /usr/bin/vikunja
```

### Systemd service

To automatically start Vikunja when your system boots and to ensure all dependent services are met, you want to use an init system like systemd.

Save the following service file to `/etc/systemd/system/vikunja.service` and adapt it to your needs:

```unit file (systemd)
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
```

If you've installed Vikunja to a directory other than `/opt/vikunja`, you need to adapt `WorkingDirectory` accordingly.

After you made all necessary modifications, it's time to start the service:

```
sudo systemctl enable vikunja
sudo systemctl start vikunja
```

### Build from source

To build vikunja from source, see [building from source]({{< ref "build-from-source.md">}}).

### Updating

[Make a backup first]({{< ref "backups.md" >}}).

Simply replace the binary with the new version, then restart Vikunja.
It will automatically run all necessary database migrations.
**Make sure to take a look at the changelog for the new version to not miss any manual steps the update may involve!**

## Docker

(Note: this assumes some familiarity with docker)

To get up and running quickly, use this command:

```
mkdir $PWD/files $PWD/db
chown 1000 $PWD/files $PWD/db
docker run -p 3456:3456 -v $PWD/files:/app/vikunja/files -v $PWD/db:/db vikunja/vikunja
```

This will expose vikunja on port `3456` on the host running the container and use sqlite as database backend.

**Note**: The container runs as the user `1000` and no group by default. 
You can use Docker's [`--user`](https://docs.docker.com/engine/reference/run/#user) flag to change that.
Make sure the new user has required permissions on the `db` and `files` folder.

You can mount a local configuration like so:

```
mkdir $PWD/files $PWD/db
chown 1000 $PWD/files $PWD/db
docker run -p 3456:3456 -v /path/to/config/on/host.yml:/app/vikunja/config.yml:ro -v $PWD/files:/app/vikunja/files -v $PWD/db:/db vikunja/vikunja
```

Though it is recommended to use environment variables or `.env` files to configure Vikunja in docker.
See [config]({{< ref "config.md">}}) for a list of available configuration options.

Check out the [docker examples]({{<ref "full-docker-example.md">}}) for more advanced configuration using mysql / postgres and a reverse proxy.

### Files volume

By default, the container stores all files uploaded and used through vikunja inside of `/app/vikunja/files` which is created as a docker volume.
You should mount the volume somewhere to the host to permanently store the files and don't lose them if the container restarts.

### Docker compose

Check out the [docker examples]({{<ref "full-docker-example.md">}}) for more advanced configuration using docker compose.

## Debian packages

Vikunja is available as deb package for installation on debian-like systems.

To install these, grab a `.deb` file from [the download page](https://dl.vikunja.io/vikunja) and run

```
dpkg -i vikunja.deb
```

This will install Vikunja to `/opt/vikunja`.
To configure it, use the config file in `/etc/vikunja/config.yml`.

## RPM

Vikunja is available as rpm package for installation on Fedora, CentOS and others.

To install these, grab a `.rpm` file from [the download page](https://dl.vikunja.io/vikunja) and run

```
rpm -i vikunja.rpm
```

To configure Vikunja, use the config file in `/etc/vikunja/config.yml`.

## FreeBSD / FreeNAS

Unfortunately, we currently can't provide pre-built binaries for FreeBSD.
As a workaround, it is possible to compile vikunja for FreeBSD directly on a FreeBSD machine, a guide is available below:

*Thanks to HungrySkeleton who originally created this guide [in the forum](https://community.vikunja.io/t/freebsd-support/69/11).*

### Jail Setup

1. Create a jail named `vikunja`
2. Set jail properties to 'auto start'
3. Mount storage (`/mnt` to `jailData/vikunja`)
4. Start jail & SSH into it

### Installing packages

```
pkg update && pkg upgrade -y
pkg install nano git go gmake
go install github.com/magefile/mage
```

### Clone vikunja repo

```
mkdir /mnt/GO/code.vikunja.io
cd /mnt/GO/code.vikunja.io
git clone https://code.vikunja.io/vikunja
cd vikunja
```

**Note:** Check out the version you want to build with `git checkout VERSION` - replace `VERSION` with the version want to use.
If you don't do this, you'll build the [latest unstable build]({{< ref "versions.md">}}), which might contain bugs.

### Compile binaries

```
cd frontend
pnpm install
pnpm run build
cd ..
mage build
```

### Create folder to install Vikunja into

```
mkdir /mnt/vikunja
cp /mnt/GO/code.vikunja.io/api/vikunja /mnt/vikunja
cd /mnt/vikunja
chmod +x /mnt/vikunja
```

### Set vikunja to boot on startup

```
nano /etc/rc.d/vikunja
```

Then paste into the file:

```
#!/bin/sh

. /etc/rc.subr

name=vikunja
rcvar=vikunja_enable

command="/mnt/vikunja/${name}"

load_rc_config $name
run_rc_command "$1"
```

Save and exit.  Then execute:

```
chmod +x /etc/rc.d/vikunja
nano /etc/rc.conf
```

Then add line to bottom of file:

```
vikunja_enable="YES"
```

Test vikunja now works with

```
service vikunja start
```

Vikunja is now available through IP:

```
192.168.1.XXX:3456
```

## Other installation resources

* [Docker Compose is MUCH Easier Than you Think - Let's Install Vikunja](https://www.youtube.com/watch?v=fGlz2PkXjuo) (Youtube)
* [Setup Vikunja using Docker Compose - Homelab Wiki](https://thehomelab.wiki/books/docker/page/setup-vikunja-using-docker-compose)
* [A Closer look at Vikunja - Email Notifications - Enable or Disable Registrations - Allow Attachments](https://www.youtube.com/watch?v=47wj9pRT6Gw) (Youtube)
* [Install Vikunja in Docker for self-hosted Task Tracking](https://smarthomepursuits.com/install-vikunja-in-docker-for-self-hosted-task-tracking/)
* [Self-Hosted To-Do List with Vikunja in Docker](https://www.youtube.com/watch?v=DqyqDWpEvKI) (Youtube)
* [Vikunja self-hosted (step by step)](https://nguyenminhhung.com/vikunja-self-hosted-step-by-step/)
* [How to Install Vikunja on Your Synology NAS](https://mariushosting.com/how-to-install-vikunja-on-your-synology-nas/)

## Configuration

See [available configuration options]({{< ref "config.md">}}).

## Default Password

After successfully installing Vikunja, there is no default user or password.
You only need to register a new account and set all the details when creating it.
