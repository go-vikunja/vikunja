# syntax=docker/dockerfile:1@sha256:b6afd42430b15f2d2a4c5a02b919e98a525b785b1aaff16747d2f623364e39b6
FROM --platform=$BUILDPLATFORM node:24.13.0-alpine@sha256:931d7d57f8c1fd0e2179dbff7cc7da4c9dd100998bc2b32afc85142d8efbc213 AS frontendbuilder

WORKDIR /build

ENV PNPM_CACHE_FOLDER=.cache/pnpm/
ENV PUPPETEER_SKIP_DOWNLOAD=true
ENV CYPRESS_INSTALL_BINARY=0

COPY frontend/pnpm-lock.yaml frontend/package.json frontend/.npmrc ./
RUN npm install -g corepack && corepack enable && \
    pnpm install --frozen-lockfile
COPY frontend/ ./
ARG RELEASE_VERSION=dev
RUN echo "{\"VERSION\": \"${RELEASE_VERSION/-g/-}\"}" > src/version.json && pnpm run build

FROM --platform=$BUILDPLATFORM ghcr.io/techknowlogick/xgo:go-1.25.x@sha256:11ac5e6cb8767caea0c62c420e053cb69554638ec255f9bbef8ed411e70c9eec AS apibuilder

RUN go install github.com/magefile/mage@latest && \
    mv /go/bin/mage /usr/local/go/bin

WORKDIR /go/src/code.vikunja.io/api
COPY . ./
COPY --from=frontendbuilder /build/dist ./frontend/dist

ARG TARGETOS TARGETARCH TARGETVARIANT RELEASE_VERSION
ENV RELEASE_VERSION=$RELEASE_VERSION

RUN export PATH=$PATH:$GOPATH/bin && \
	mage build:clean && \
    mage release:xgo "${TARGETOS}/${TARGETARCH}/${TARGETVARIANT}"

RUN mkdir -p /tmp && chmod 1777 /tmp

#  ┬─┐┬ ┐┌┐┐┌┐┐┬─┐┬─┐
#  │┬┘│ │││││││├─ │┬┘
#  ┘└┘┘─┘┘└┘┘└┘┴─┘┘└┘

# The actual image
FROM scratch

LABEL org.opencontainers.image.url='https://kanban.yarlis.com'
LABEL org.opencontainers.image.source='https://github.com/siri1410/projectos-engine'
LABEL org.opencontainers.image.licenses='AGPL-3.0-or-later'
LABEL org.opencontainers.image.title='ProjectOS'

WORKDIR /app/projectos
ENTRYPOINT [ "/app/projectos/projectos" ]
EXPOSE 3456

COPY --from=apibuilder --chown=1000:1000 /tmp /tmp

USER 1000

# White-label: the engine reads the PROJECTOS_ config prefix (see pkg/config/config.go).
ENV PROJECTOS_SERVICE_ROOTPATH=/app/projectos/
ENV PROJECTOS_DATABASE_PATH=/db/projectos.db

COPY --from=apibuilder /build/vikunja-* projectos
COPY --from=apibuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
