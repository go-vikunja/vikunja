# Stage 1: Build application
FROM node:16 AS compile-image

WORKDIR /build

ARG USE_RELEASE=false
ARG RELEASE_VERSION=main

ENV YARN_CACHE_FOLDER .cache/yarn/
COPY .  ./

RUN \
  if [ $USE_RELEASE = true ]; then \
    rm -rf dist/ && \
    wget https://dl.vikunja.io/frontend/vikunja-frontend-$RELEASE_VERSION.zip -O frontend-release.zip && \
    unzip frontend-release.zip -d dist/ && \
    exit 0; \
  fi && \
  # Build the frontend
  yarn install --frozen-lockfile --network-timeout 100000 && \
  echo '{"VERSION": "'$(git describe --tags --always --abbrev=10 | sed 's/-/+/' | sed 's/^v//' | sed 's/-g/-/')'"}' > src/version.json && \
  yarn run build

# Stage 2: copy 
FROM nginx

COPY nginx.conf /etc/nginx/nginx.conf
COPY run.sh /run.sh

# copy compiled files from stage 1
COPY --from=compile-image /build/dist /usr/share/nginx/html

# Unprivileged user
ENV PUID 1000
ENV PGID 1000

LABEL maintainer="maintainers@vikunja.io"

CMD "/run.sh"
