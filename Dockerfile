
###################################
#Build stage
FROM golang:1.11-alpine AS build-env

ARG VIKUNJA_VERSION
ENV TAGS "sqlite"
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

#Build deps
RUN apk --no-cache add build-base git

#Setup repo
COPY . ${GOPATH}/src/code.vikunja.io/api
WORKDIR ${GOPATH}/src/code.vikunja.io/api

#Checkout version if set
RUN if [ -n "${VIKUNJA_VERSION}" ]; then git checkout "${VIKUNJA_VERSION}"; fi \
 && make clean build

FROM alpine:3.7
LABEL maintainer="maintainers@vikunja.io"

EXPOSE 3456

RUN apk --no-cache add \
    bash \
    ca-certificates \
    curl \
    gettext \
    linux-pam \
    s6 \
    sqlite \
    su-exec \
    tzdata

COPY docker /
COPY --from=build-env /go/src/code.vikunja.io/api/public /app/vikunja/public
COPY --from=build-env /go/src/code.vikunja.io/api/templates /app/vikunja/templates
COPY --from=build-env /go/src/code.vikunja.io/api/vikunja /app/vikunja/vikunja

ENV VIKUNJA_SERVICE_ROOTPATH=/app/vikunja/

ENTRYPOINT ["/bin/s6-svscan", "/etc/services.d"]
CMD []
