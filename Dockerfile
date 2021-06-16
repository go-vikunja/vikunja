
##############
# Build stage
FROM golang:1-alpine3.12 AS build-env

ARG VIKUNJA_VERSION
ENV TAGS "sqlite"
ENV GO111MODULE=on

# Build deps
RUN apk --no-cache add build-base git

# Setup repo
COPY . ${GOPATH}/src/code.vikunja.io/api
WORKDIR ${GOPATH}/src/code.vikunja.io/api

# Checkout version if set
RUN if [ -n "${VIKUNJA_VERSION}" ]; then git checkout "${VIKUNJA_VERSION}"; fi \
 && go install github.com/magefile/mage \
 && mage build:clean build

###################
# The actual image
# Note: I wanted to use the scratch image here, but unfortunatly the go-sqlite bindings require cgo and
# because of this, the container would not start when I compiled the image without cgo.
FROM alpine:3.14
LABEL maintainer="maintainers@vikunja.io"

WORKDIR /app/vikunja/
COPY --from=build-env /go/src/code.vikunja.io/api/vikunja .
ENV VIKUNJA_SERVICE_ROOTPATH=/app/vikunja/

# Dynamic permission changing stuff
ENV PUID 1000
ENV PGID 1000
RUN apk --no-cache add shadow && \
  addgroup -g ${PGID} vikunja && \
  adduser -s /bin/sh -D -G vikunja -u ${PUID} vikunja -h /app/vikunja -H && \
  chown vikunja -R /app/vikunja
COPY run.sh /run.sh

# Fix time zone settings not working
RUN apk --no-cache add tzdata

# Files permissions
RUN mkdir /app/vikunja/files && \
  chown -R vikunja /app/vikunja/files
VOLUME /app/vikunja/files

CMD ["/run.sh"]
EXPOSE 3456
