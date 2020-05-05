# Stage 1: Build application
FROM node:13.14.0 AS compile-image

WORKDIR /build

COPY .  ./

RUN \
  # Build the frontend
  yarn install --frozen-lockfile && \
  yarn run build && \
  # Override config
  sed -i 's/http\:\/\/localhost\:8080\/api\/v1/\/api\/v1/g' dist/index.html

# Stage 2: copy 
FROM nginx

RUN apt-get update && apt-get install -y apt-utils openssl && \
  mkdir -p /etc/nginx/ssl && \
  openssl genrsa -out /etc/nginx/ssl/dummy.key 2048 && \
  openssl req -new -key /etc/nginx/ssl/dummy.key -out /etc/nginx/ssl/dummy.csr -subj "/C=DE/L=Berlin/O=Vikunja/CN=Vikunja Snakeoil" && \
  openssl x509 -req -days 3650 -in /etc/nginx/ssl/dummy.csr -signkey /etc/nginx/ssl/dummy.key -out /etc/nginx/ssl/dummy.crt

COPY nginx.conf /etc/nginx/nginx.conf

# copy compiled files from stage 1
COPY --from=compile-image /build/dist /usr/share/nginx/html

LABEL maintainer="maintainers@vikunja.io"