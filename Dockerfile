# Stage 1: Build application
FROM node:13.14.0 AS compile-image

WORKDIR /build

ENV YARN_CACHE_FOLDER .cache/yarn/
COPY .  ./

RUN \
  # Build the frontend
  yarn install --frozen-lockfile --network-timeout 100000 && \
  echo '{"VERSION": "'$(git describe --tags --always --abbrev=10 | sed 's/-/+/' | sed 's/^v//' | sed 's/-g/-/')'"}' > src/version.json && \
  yarn run build

# Stage 2: copy 
FROM nginx

RUN apt-get update && apt-get install -y apt-utils openssl && \
  mkdir -p /etc/nginx/ssl && \
  openssl genrsa -out /etc/nginx/ssl/dummy.key 2048 && \
  openssl req -new -key /etc/nginx/ssl/dummy.key -out /etc/nginx/ssl/dummy.csr -subj "/C=DE/L=Berlin/O=Vikunja/CN=Vikunja Snakeoil" && \
  openssl x509 -req -days 3650 -in /etc/nginx/ssl/dummy.csr -signkey /etc/nginx/ssl/dummy.key -out /etc/nginx/ssl/dummy.crt

COPY nginx.conf /etc/nginx/nginx.conf
COPY run.sh /run.sh

# copy compiled files from stage 1
COPY --from=compile-image /build/dist /usr/share/nginx/html

# Unprivileged user
ENV PUID 1000
ENV PGID 1000

LABEL maintainer="maintainers@vikunja.io"

CMD "/run.sh"