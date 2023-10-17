---
date: "2019-02-12:00:00+02:00"
title: "Reverse Proxy"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Setup behind a reverse proxy which also serves the frontend

These examples assume you have an instance of the backend running on your server listening on port `3456`.
If you've changed this setting, you need to update the server configurations accordingly.

{{< table_of_contents >}}

## NGINX

Below are two example configurations which you can put in your `nginx.conf`:

You may need to adjust `server_name` and `root` accordingly.

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
    
    location ~* ^/(api|dav|\.well-known)/ {
        proxy_pass http://localhost:3456;
        client_max_body_size 20M;
    }
}
{{< /highlight >}}

<div class="notification is-warning">
<b>NOTE:</b> If you change the max upload size in Vikunja's settings, you'll need to also change the <code>client_max_body_size</code> in the nginx proxy config.
</div>

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
    
    location ~* ^/(api|dav|\.well-known)/ {
        proxy_pass http://localhost:3456;
        client_max_body_size 20M;
    }
}
{{< /highlight >}}

<div class="notification is-warning">
<b>NOTE:</b> If you change the max upload size in Vikunja's settings, you'll need to also change the <code>client_max_body_size</code> in the nginx proxy config.
</div>

## NGINX Proxy Manager (NPM)

### Method 1

Following the [Docker Walkthrough]({{< ref "docker-start-to-finish.md" >}}) guide, you should be able to get Vikunja to work via HTTP connection to your server IP.

From there, all you have to do is adjust the following things:

#### In `docker-compose.yml`

Under `api:`, 

1. Change `VIKUNJA_SERVICE_FRONTENDURL:` to your desired domain with `https://` and `/`.

2. Expose your desired port on host under `ports:`.

example:

```yaml
  api:
    image: vikunja/api
    environment:
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: secret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_DATABASE_DATABASE: vikunja
      VIKUNJA_SERVICE_JWTSECRET: <your-random-secret>
      VIKUNJA_SERVICE_FRONTENDURL: https://vikunja.your-domain.com/ # change vikunja.your-domain.com to your desired domain/subdomain.
    ports:
      - 3456:3456 # Change 3456 on the left to the port of your choice.
    volumes: 
      - ./files:/app/vikunja/files
    depends_on:
      - db
    restart: unless-stopped
```

Under `frontend:`,

1. Add `VIKUNJA_API_URL:` under `environment:` and input your desired `API` domain with `https://` and `/api/v1/`. The `API` domain should be different from the one in `VIKUNJA_SERVICE_FRONTENDURL:`.

example:

```yaml
frontend:
    image: vikunja/frontend
    environment:
      VIKUNJA_API_URL: https://api.your-domain.com/api/v1/ # change api.your-domain.com to your desired domain/subdomain, it should be different from your frontend domain
    restart: unless-stopped
```

Under `proxy:`,

1. Since we'll be using Nginx Proxy Manager, it should by default uses the port `80` and thus you should change `ports:` to expose another port not occupied by any service.

example:

```yaml
  proxy:
      image: nginx
      ports: 
        - 1078:80 # change the number infront (host port) to whatever you desire, but make sure it's not 80 which will be used by Nginx Proxy Manager
      volumes:
        - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
      depends_on:
        - api
        - frontend
      restart: unless-stopped
```

#### In your DNS provider

Add two `A` records that points to your server IP. 

1. `vikunja` for accessing the frontend

2. `api` for accessing the api

You are of course free to change them to whatever domain/subdomain you desire and modify the `docker-compose.yml` accordingly but the two should be different.

(Tested on Cloudflare DNS. Settings are different for different DNS provider, in this case the end result should bei `vikunja.your-domain.com` and `api.your-domain.com` respectively.)

#### In Nginx Proxy Manager

Add two Proxy Host as you normally would, and you don't have to add anything extra in Advanced.

##### Frontend

Under `Details`:

```
Domain Names:
    vikunja.your-domain.com
Scheme:
    http
Forward Hostname/IP:
    your-server-ip
Forward Port:
    1078
Cached Assets:
    Optional.
Block Common Exploits:
    Toggled.
Websockets Support:
    Toggled.
```

Under `SSL`:

```
SSL Certificate:
    However you prefer.
Force SSL:
    Toggled.
HTTP/2 Support:
    Toggled.
HSTS Enabled:
    Toggled.
HSTS Subdomains:
    Toggled.
Use a DNS Challenge:
    Not toggled.
Email Address for Let's Encrypt:
    your-email@email.com
```

##### API

Under `Details`:

```
Domain Names:
    api.your-domain.com
Scheme:
    http
Forward Hostname/IP:
    your-server-ip
Forward Port:
    3456
Cached Assets:
    Optional.
Block Common Exploits:
    Toggled.
Websockets Support:
    Toggled.
```

Under `SSL`:

```
SSL Certificate:
    However you prefer.
Force SSL:
    Toggled.
HTTP/2 Support:
    Toggled.
HSTS Enabled:
    Toggled.
HSTS Subdomains:
    Toggled.
Use a DNS Challenge:
    Not toggled.
Email Address for Let's Encrypt:
    your-email@email.com
```

Your Vikunja service should now work and your HTTPS frontend should be able to reach the API after `docker-compose`.

### Method 2

1. Create a standard Proxy Host for the Vikunja Frontend within NPM and point it to the URL you plan to use. The next several steps will enable the Proxy Host to successfully navigate to the API (on port 3456).
2. Verify that the page will pull up in your browser. (Do not bother trying to log in. It won't work. Trust me.)
3. Now, we'll work with the NPM container, so you need to identify the container name for your NPM installation. e.g. NGINX-PM
4. From the command line, enter `sudo docker exec -it [NGINX-PM container name] /bin/bash` and navigate to the proxy hosts folder where the `.conf` files are stashed. Probably `/data/nginx/proxy_host`. (This folder is a persistent folder created in the NPM container and mounted by NPM.)
5. Locate the `.conf` file where the server_name inside the file matches your Vikunja Proxy Host. Once found, add the following code, unchanged, just above the existing location block in that file. (They are listed by number, not name.)
```nginx
location ~* ^/(api|dav|\.well-known)/ {
		proxy_pass http://api:3456;
		client_max_body_size 20M;
	}
```
6. After saving the edited file, return to NPM's UI browser window and refresh the page to verify your Proxy Host for Vikunja is still online.
7. Now, switch over to your Vikunja browser window and hit refresh. If you configured your URL correctly in original Vikunja container, you should be all set and the browser will correctly show Vikunja. If not, you'll need to adjust the address in the top of the login subscreen to match your proxy address.

## Apache

Put the following config in `cat /etc/apache2/sites-available/vikunja.conf`:

{{< highlight aconf >}}
<VirtualHost *:80>
    ServerName localhost
   
    <Proxy *>
      Order Deny,Allow
      Allow from all
    </Proxy>
    ProxyPass /api http://localhost:3456/api
    ProxyPassReverse /api http://localhost:3456/api
    ProxyPass /dav http://localhost:3456/dav
    ProxyPassReverse /dav http://localhost:3456/dav
    ProxyPass /.well-known http://localhost:3456/.well-known
    ProxyPassReverse /.well-known http://localhost:3456/.well-known

    DocumentRoot /var/www/html
    RewriteEngine On
 	RewriteRule ^\/?(favicon\.ico|assets|audio|fonts|images|manifest\.webmanifest|robots\.txt|sw\.js|workbox-.*|api|dav|\.well-known) - [L]
    RewriteRule ^(.*)$ /index.html [QSA,L]
</VirtualHost>
{{< /highlight >}}

**Note:** The apache modules `proxy`, `proxy_http` and `rewrite` must be enabled for this.

For more details see the [frontend apache configuration]({{< ref "install-frontend.md#apache">}}).

## Caddy

{{< highlight conf >}}
vikunja.domainname.tld {
	@paths {
		path /api/* /.well-known/* /dav/*
	}
	handle @paths {
		reverse_proxy 127.0.0.1:3456
	}

	handle {
		encode zstd gzip
		root * /var/www/html/vikunja
		try_files {path} index.html
		file_server
	}
}
{{< /highlight >}}
