
###################################
#Build stage
FROM golang:1.11-alpine3.7 AS build-env

ARG VIKUNJA_VERSION
ENV TAGS "sqlite"

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
COPY --from=build-env /go/src/code.vikunja.io/api/public /app/vikunja/
COPY --from=build-env /go/src/code.vikunja.io/api/vikunja /app/vikunja/vikunja

ENTRYPOINT ["/bin/s6-svscan", "/etc/services.d"]
CMD []
