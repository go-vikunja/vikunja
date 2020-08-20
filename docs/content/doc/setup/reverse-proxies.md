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
    RewriteRule ^\/?(config\.json|favicon\.ico|css|fonts|images|img|js|api|dav|\.well-known) - [L]
    RewriteRule ^(.*)$ /index.html [QSA,L]
</VirtualHost>
{{< /highlight >}}

**Note:** The apache modules `proxy`, `proxy_http` and `rewrite` must be enabled for this.

For more details see the [frontend apache configuration]({{< ref "install-frontend.md#apache">}}).