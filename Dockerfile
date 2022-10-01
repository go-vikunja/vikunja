
##############
# Build stage
FROM golang:1.19-alpine AS build-env

RUN apk --no-cache add build-base git && \
  go install github.com/magefile/mage@latest && \
  mv /go/bin/mage /usr/local/go/bin

ARG VIKUNJA_VERSION

# Setup repo
COPY . /go/src/code.vikunja.io/api
WORKDIR /go/src/code.vikunja.io/api

# Checkout version if set
RUN if [ -n "${VIKUNJA_VERSION}" ]; then git checkout "${VIKUNJA_VERSION}"; fi \
 && mage build:clean build

###################
# The actual image
# Note: I wanted to use the scratch image here, but unfortunatly the go-sqlite bindings require cgo and
# because of this, the container would not start when I compiled the image without cgo.
FROM alpine:3.16
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

# Add time zone data
RUN apk --no-cache add tzdata

# Files permissions
RUN mkdir /app/vikunja/files && \
  chown -R vikunja /app/vikunja/files
VOLUME /app/vikunja/files

CMD ["/run.sh"]
EXPOSE 3456
