
###################################
#Build stage
FROM golang:1.11-alpine3.7 AS build-env

ARG VIKUNJA_VERSION
ENV TAGS "sqlite"

#Build deps
RUN apk --no-cache add build-base git

#Setup repo
COPY . ${GOPATH}/src/code.vikunja.io/vikunja
WORKDIR ${GOPATH}/src/code.vikunja.io/vikunja

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

ENTRYPOINT ["/usr/bin/entrypoint"]
CMD ["/bin/s6-svscan", "/etc/s6"]

COPY docker /
COPY --from=build-env /go/src/code.vikunja.io/vikunja/vikunja /app/vikunja/vikunja
RUN ln -s /app/vikunja/vikunja /usr/local/bin/vikunja
