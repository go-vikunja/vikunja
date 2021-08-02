---
date: "2019-02-12:00:00+02:00"
title: "Install Frontend"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Frontend

Installing the frontend is just a matter of hosting a bunch of static files somewhere.

With nginx or apache, you have to [download](https://vikunja.io/en/download/) the frontend files first.
Unzip them and store them somewhere your server can access them.

You also need to configure a rewrite condition to internally redirect all requests to `index.html` which handles all urls. 

{{< table_of_contents >}}

## API URL configuration

By default, the frontend assumes it can reach the api at `/api/v1` relative to the frontend url.
This means that if you make the frontend available at, say `https://vikunja.example.com`, it tries to reach the api
at `https://vikunja.example.com/api/v1`.
In this scenario it is not possible for the frontend and the api to live on seperate servers or even just seperate 
ports on the same server with [the use of a reverse proxy]({{< ref "reverse-proxies.md">}}).

To make configurations like this possible, the api url can be set in the `index.html` file of the frontend releases.
Just open the file with a text editor - there are comments which will explain how to set the url.

**Note:** This needs to be done again after every update. 
(If you have a good idea for a better solution than this, we'd love to [hear it](https://vikunja.io/contact/))

## Docker

The docker image is based on nginx and just contains all nessecary files for the frontend.

To run it, all you need is

{{< highlight bash >}}
docker run -p 80:80 vikunja/frontend
{{< /highlight >}}

which will run the docker image and expose port 80 on the host.

See [full docker example]({{< ref "full-docker-example.md">}}) for more varations of this config.

### Setting user and group id of the user running vikunja

You can set the user and group id of the user running vikunja with the `PUID` and `PGID` evironment variables.
This follows the pattern used by [the linuxserver.io](https://docs.linuxserver.io/general/understanding-puid-and-pgid) docker images.

### API URL configuration in docker

When running the frontend with docker, it is possible to set the environment variable `$VIKUNJA_API_URL` to the api url.
It is therefore not needed to change the url manually inside the docker container.

## NGINX

Below are two example configurations which you can put in your `nginx.conf`:

You may need to adjust `server_name` and `root` accordingly.

After configuring them, you need to reload nginx (`service nginx reload`).

### with gzip enabled (recommended)

{{< highlight conf >}}
gzip  on;
gzip_disable "msie6";

gzip_vary on;
gzip_proxied any;
gzip_comp_level 6;
gzip_buffers 16 8k;
gzip_http_version 1.1;
gzip_min_length 256;
gzip_types text/plain text/css application/json application/x-javascript text/xml application/xml application/xml+rss text/javascript application/vnd.ms-fontobject application/x-font-ttf font/opentype image/svg+xml;

server {
    listen       80;
    server_name  localhost;

    location / {
        root   /path/to/vikunja/static/frontend/files;
        try_files $uri $uri/ /;
        index  index.html index.htm;
    }
}
{{< /highlight >}}

### without gzip

{{< highlight conf >}}
server {
    listen       80;
    server_name  localhost;

    location / {
        root   /path/to/vikunja/static/frontend/files;
        try_files $uri $uri/ /;
        index  index.html index.htm;
    }
}
{{< /highlight >}}

## Apache

Apache needs to have `mod_rewrite` enabled for this to work properly:

{{< highlight bash >}}
a2enmod rewrite
service apache2 restart
{{< /highlight >}}

Put the following config in `cat /etc/apache2/sites-available/vikunja.conf`:

{{< highlight aconf >}}
<VirtualHost *:80>
    ServerName localhost
    DocumentRoot /path/to/vikunja/static/frontend/files
    RewriteEngine On
 	RewriteRule ^\/?(favicon\.ico|assets|audio|fonts|images|manifest\.webmanifest|robots\.txt|sw\.js|workbox-.*|api|dav|\.well-known) - [L]
    RewriteRule ^(.*)$ /index.html [QSA,L]
</VirtualHost>
{{< /highlight >}}

You probably want to adjust `ServerName` and `DocumentRoot`.

Once you've customized your config, you need to enable it:

{{< highlight bash >}}
a2ensite vikunja
service apache2 reload
{{< /highlight >}}

## Updating

To update, it should be enough to download the new files and overwrite the old ones.
The paths contain hashes, so all caches are invalidated automatically.
