# Stage 1: Build application
FROM node:18-alpine AS compile-image

WORKDIR /build

ARG USE_RELEASE=false
ARG RELEASE_VERSION=main

RUN \
  if [ $USE_RELEASE = true ]; then \
    wget https://dl.vikunja.io/frontend/vikunja-frontend-$RELEASE_VERSION.zip -O frontend-release.zip && \
    unzip frontend-release.zip -d dist/ && \
    exit 0; \
  fi

ENV PNPM_CACHE_FOLDER .cache/pnpm/

# pnpm fetch does require only lockfile
COPY pnpm-lock.yaml ./

RUN \
  # https://pnpm.io/installation#using-corepack
  corepack enable && \
  corepack prepare pnpm@7.9.3 --activate && \
  # Build the frontend
	pnpm fetch

ADD .  ./

RUN apk add --no-cache git

RUN \
	pnpm install --offline && \
	echo '{"VERSION": "'$(git describe --tags --always --abbrev=10 | sed 's/-/+/' | sed 's/^v//' | sed 's/-g/-/')'"}' > src/version.json && \
	pnpm run build

# Stage 2: copy 
FROM nginx:alpine

COPY nginx.conf /etc/nginx/nginx.conf
COPY run.sh /run.sh

# copy compiled files from stage 1
COPY --from=compile-image /build/dist /usr/share/nginx/html

# Unprivileged user
ENV PUID 1000
ENV PGID 1000

LABEL maintainer="maintainers@vikunja.io"

RUN apk add --no-cache \
  # for sh file
	bash \
	# installs usermod and groupmod
	shadow

CMD "/run.sh"
