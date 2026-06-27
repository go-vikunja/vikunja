# syntax=docker/dockerfile:1@sha256:87999aa3d42bdc6bea60565083ee17e86d1f3339802f543c0d03998580f9cb89
FROM --platform=$BUILDPLATFORM node:24.13.0-alpine@sha256:cd6fb7efa6490f039f3471a189214d5f548c11df1ff9e5b181aa49e22c14383e AS frontendbuilder

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
    (cd build && mage release:xgo vikunja "${TARGETOS}/${TARGETARCH}/${TARGETVARIANT}")

RUN mkdir -p /tmp && chmod 1777 /tmp

#  ┬─┐┬ ┐┌┐┐┌┐┐┬─┐┬─┐
#  │┬┘│ │││││││├─ │┬┘
#  ┘└┘┘─┘┘└┘┘└┘┴─┘┘└┘

# The actual image
FROM scratch

LABEL org.opencontainers.image.authors='maintainers@vikunja.io'
LABEL org.opencontainers.image.url='https://vikunja.io'
LABEL org.opencontainers.image.documentation='https://vikunja.io/docs'
LABEL org.opencontainers.image.source='https://code.vikunja.io/vikunja'
LABEL org.opencontainers.image.licenses='AGPLv3'
LABEL org.opencontainers.image.title='Vikunja'

WORKDIR /app/vikunja
ENTRYPOINT [ "/app/vikunja/vikunja" ]
EXPOSE 3456

COPY --from=apibuilder --chown=1000:1000 --chmod=1777 /tmp /tmp

USER 1000

ENV VIKUNJA_SERVICE_ROOTPATH=/app/vikunja/
ENV VIKUNJA_DATABASE_PATH=/db/vikunja.db

COPY --from=apibuilder /build/vikunja-* vikunja
COPY --from=apibuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
