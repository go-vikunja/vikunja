FROM nginx

MAINTAINER maintainers@vikunja.io

RUN apt-get update && apt-get install -y apt-utils openssl && \
  mkdir -p /etc/nginx/ssl && \
  openssl genrsa -out /etc/nginx/ssl/dummy.key 2048 && \
  openssl req -new -key /etc/nginx/ssl/dummy.key -out /etc/nginx/ssl/dummy.csr -subj "/C=DE/L=Berlin/O=Vikunja/CN=Vikunja Snakeoil" && \
  openssl x509 -req -days 3650 -in /etc/nginx/ssl/dummy.csr -signkey /etc/nginx/ssl/dummy.key -out /etc/nginx/ssl/dummy.crt

ADD nginx.conf /etc/nginx/nginx.conf

COPY dist /usr/share/nginx/html