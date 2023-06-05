# syntax=docker/dockerfile:1
#  ┬─┐┬ ┐o┬  ┬─┐
#  │─││ │││  │ │
#  ┘─┘┘─┘┘┘─┘┘─┘

FROM --platform=$BUILDPLATFORM techknowlogick/xgo:go-1.20.x AS builder

RUN go install github.com/magefile/mage@latest && \
    mv /go/bin/mage /usr/local/go/bin

WORKDIR /go/src/code.vikunja.io/api
COPY . ./

ARG TARGETOS TARGETARCH TARGETVARIANT

RUN export PATH=$PATH:$GOPATH/bin && \
	mage build:clean && \
    mage release:xgo "${TARGETOS}/${TARGETARCH}/${TARGETVARIANT}"

#  ┬─┐┬ ┐┌┐┐┌┐┐┬─┐┬─┐
#  │┬┘│ │││││││├─ │┬┘
#  ┘└┘┘─┘┘└┘┘└┘┴─┘┘└┘

# The actual image
# Note: I wanted to use the scratch image here, but unfortunatly the go-sqlite bindings require cgo and
# because of this, the container would not start when I compiled the image without cgo.
FROM alpine:3.18 AS runner
LABEL maintainer="maintainers@vikunja.io"
WORKDIR /app/vikunja
ENTRYPOINT [ "/sbin/tini", "-g", "--", "/entrypoint.sh" ]
EXPOSE 3456

ENV VIKUNJA_SERVICE_ROOTPATH=/app/vikunja/
ENV PUID 1000
ENV PGID 1000

RUN apk --update --no-cache add tzdata tini shadow && \
    addgroup vikunja --gid "$PGID" && \
    adduser -s /bin/sh -D -G vikunja vikunja --uid "$PUID" -h /app/vikunja -H
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod 0755 /entrypoint.sh && mkdir files

COPY --from=builder /build/vikunja-* vikunja
