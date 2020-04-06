
##############
# Build stage
FROM golang:1-alpine AS build-env

ARG VIKUNJA_VERSION
ENV TAGS "sqlite"
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

# Build deps
RUN apk --no-cache add build-base git

# Setup repo
COPY . ${GOPATH}/src/code.vikunja.io/api
WORKDIR ${GOPATH}/src/code.vikunja.io/api

# Checkout version if set
RUN if [ -n "${VIKUNJA_VERSION}" ]; then git checkout "${VIKUNJA_VERSION}"; fi \
 && make clean generate build

###################
# The actual image
# Note: I wanted to use the scratch image here, but unfortunatly the go-sqlite bindings require cgo and
# because of this, the container would not start when I compiled the image without cgo.
FROM alpine:3.11
LABEL maintainer="maintainers@vikunja.io"

WORKDIR /app/vikunja/
COPY --from=build-env /go/src/code.vikunja.io/api/vikunja .
RUN adduser -S -D vikunja -h /app/vikunja -H \
  && chown vikunja -R /app/vikunja
ENV VIKUNJA_SERVICE_ROOTPATH=/app/vikunja/

# Fix time zone settings not working
RUN apk --no-cache add tzdata

# Files permissions
RUN mkdir /app/vikunja/files && \
  chown -R vikunja /app/vikunja/files
VOLUME /app/vikunja/files

USER vikunja
CMD ["/app/vikunja/vikunja"]
EXPOSE 3456
