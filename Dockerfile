
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

#Checkout version if set
RUN if [ -n "${VIKUNJA_VERSION}" ]; then git checkout "${VIKUNJA_VERSION}"; fi \
 && make clean generate build

###################
# The actual image
# Note: I wanted to use the scratch image here, but unfortunatly the go-sqlite bindings require cgo and
# for whatever reason, the container would not start when I compiled the image without cgo.
FROM alpine:3.9
LABEL maintainer="maintainers@vikunja.io"

WORKDIR /app/vikunja/
COPY --from=build-env /go/src/code.vikunja.io/api/vikunja .
RUN chown nobody:nogroup -R /app/vikunja
ENV VIKUNJA_SERVICE_ROOTPATH=/app/vikunja/

USER nobody:nogroup
CMD ["/app/vikunja/vikunja"]
EXPOSE 3456
