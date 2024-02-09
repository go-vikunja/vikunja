---
date: "2019-02-12:00:00+02:00"
title: "Reverse Proxy"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Setup behind a reverse proxy

These examples assume you have an instance of Vikunja running on your server listening on port `3456`.
If you've changed this setting, you need to update the server configurations accordingly.

{{< table_of_contents >}}

## NGINX

You may need to adjust `server_name` and `root` accordingly.

```conf
server {
    listen       80;
    server_name  localhost;

    location / {
        proxy_pass http://localhost:3456;
        client_max_body_size 20M;
    }
}
```

<div class="notification is-warning">
<b>NOTE:</b> If you change the max upload size in Vikunja's settings, you'll need to also change the <code>client_max_body_size</code> in the nginx proxy config.
</div>

## NGINX Proxy Manager (NPM)

Following the [Docker Walkthrough]({{< ref "docker-start-to-finish.md" >}}) guide, you should be able to get Vikunja to work via HTTP connection to your server IP.

From there, all you have to do is adjust the following things:

### In `docker-compose.yml`

1. Change `VIKUNJA_SERVICE_PUBLICURL:` to your desired domain with `https://` and `/`.
2. Expose your desired port on host under `ports:`.

example:

```yaml
  vikunja:
    image: vikunja/vikunja
    environment:
      VIKUNJA_SERVICE_PUBLICURL: https://vikunja.your-domain.com/ # change vikunja.your-domain.com to your desired domain/subdomain.
      VIKUNJA_DATABASE_HOST: db
      VIKUNJA_DATABASE_PASSWORD: secret
      VIKUNJA_DATABASE_TYPE: mysql
      VIKUNJA_DATABASE_USER: vikunja
      VIKUNJA_DATABASE_DATABASE: vikunja
      VIKUNJA_SERVICE_JWTSECRET: <your-random-secret>
    ports:
      - 3456:3456 # Change 3456 on the left to the port of your choice.
    volumes: 
      - ./files:/app/vikunja/files
    depends_on:
      - db
    restart: unless-stopped
```

### In your DNS provider

Add an `A` records that points to your server IP.

You are of course free to change them to whatever domain/subdomain you desire and modify the `docker-compose.yml` accordingly.

(Tested on Cloudflare DNS. Settings are different for different DNS provider, in this case the end result should be `vikunja.your-domain.com`)

### In Nginx Proxy Manager

Add a Proxy Host as you normally would, and you don't have to add anything extra in Advanced.

Under `Details`:

```
Domain Names:
    vikunja.your-domain.com
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

## Apache

Put the following config in `cat /etc/apache2/sites-available/vikunja.conf`:

```aconf
<VirtualHost *:80>
    ServerName localhost
   
    <Proxy *>
      Order Deny,Allow
      Allow from all
    </Proxy>
    ProxyPass / http://localhost:3456/
    ProxyPassReverse / http://localhost:3456/
</VirtualHost>
```

**Note:** The apache modules `proxy`, `proxy_http` and `rewrite` must be enabled for this.

## Caddy

Use the following Caddyfile to get Vikunja up and running:

```conf
vikunja.domainname.tld {
	handle /* {
		reverse_proxy 127.0.0.1:3456
	}
}
```
