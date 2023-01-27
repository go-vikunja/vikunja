# syntax=docker/dockerfile:1
#  ┬─┐┬ ┐o┬  ┬─┐
#  │─││ │││  │ │
#  ┘─┘┘─┘┘┘─┘┘─┘

FROM --platform=$BUILDPLATFORM node:18-alpine AS builder

WORKDIR /build

ARG USE_RELEASE=false
ARG RELEASE_VERSION=main
ENV PNPM_CACHE_FOLDER .cache/pnpm/

COPY package.json ./
COPY pnpm-lock.yaml ./

RUN if [ "$USE_RELEASE" != true ]; then \
      # https://pnpm.io/installation#using-corepack
      corepack enable && \
      pnpm install; \
    fi

COPY . ./

RUN if [ "$USE_RELEASE" != true ]; then \
      apk add --no-cache --virtual .build-deps git jq && \
      git describe --tags --always --abbrev=10 | sed 's/-/+/; s/^v//; s/-g/-/' | \
      xargs -0 -I{} jq -Mcnr --arg version {} '{VERSION:$version}' | \
      tee src/version.json && \
      apk del .build-deps; \
    fi

RUN if [ "$USE_RELEASE" = true ]; then \
      wget "https://dl.vikunja.io/frontend/vikunja-frontend-${RELEASE_VERSION}.zip" -O frontend-release.zip && \
      unzip frontend-release.zip -d dist/; \
    else \
      # we don't use corepack prepare here by intend since
      # we have renovate to keep our dependencies up to date
      # Build the frontend
      pnpm run build; \
  fi

#  ┌┐┐┌─┐o┌┐┐┐ │
#  ││││ ┬││││┌┼┘
#  ┘└┘┘─┘┘┘└┘┘ └

FROM nginx:stable-alpine AS runner
WORKDIR /usr/share/nginx/html
LABEL maintainer="maintainers@vikunja.io"

ENV VIKUNJA_HTTP_PORT 80
ENV VIKUNJA_HTTP2_PORT 81
ENV VIKUNJA_LOG_FORMAT main
ENV VIKUNJA_API_URL /api/v1
ENV VIKUNJA_SENTRY_ENABLED false
ENV VIKUNJA_SENTRY_DSN https://85694a2d757547cbbc90cd4b55c5a18d@o1047380.ingest.sentry.io/6024480

COPY docker/injector.sh /docker-entrypoint.d/50-injector.sh
COPY docker/ipv6-disable.sh /docker-entrypoint.d/60-ipv6-disable.sh
COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY docker/templates/. /etc/nginx/templates/
# copy compiled files from stage 1
COPY --from=builder /build/dist ./
# manage permissions
RUN chmod 0755 /docker-entrypoint.d/*.sh /etc/nginx/templates && \
    chmod -R 0644 /etc/nginx/nginx.conf && \
    chown -R nginx:nginx ./ /etc/nginx/conf.d /etc/nginx/templates && \
    rm -f /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
# unprivileged user
USER nginx
