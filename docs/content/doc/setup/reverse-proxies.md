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

1. Create a standard Proxy Host for the Vikunja Frontend within NPM and point it to the URL you plan to use. The next several steps will enable the Proxy Host to successfully navigate to the API (on port 3456).
2. Verify that the page will pull up in your browser. (Do not bother trying to log in. It won't work. Trust me.)
3. Go back and edit the proxy host for Vikunja, and this time go to the Custom Locations tab, and add the following.location bar to enter  
```
 ~* ^/(api|dav|\.well-known)/ {
```
 Then click the settings wheel next to the following
 ```
client_max_body_size 20M;
```
It should end up looking like this.
Replace "api" with the IP address of your API container if needed.
![image](https://user-images.githubusercontent.com/724777/169915063-40bb6b0a-4e92-40b7-b089-beb98d4ebc0a.png)

Finally, click save.

4. Now, switch over to your Vikunja browswer window and hit refresh. If you configured your URL correctly in original Vikunja container, you should be all set and the browser will correctly show Vikunja. If not, you'll need to adjust the address in the top of the login subscreen to match your proxy address.

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
