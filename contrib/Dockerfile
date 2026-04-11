FROM golang:1.26-alpine AS build

RUN apk add --no-cache \
  gcc \
  git \
  musl-dev

RUN git clone https://github.com/cblgh/cerca /build

WORKDIR /build

RUN CGO_ENABLED=1 go build -v ./cmd/cerca

FROM alpine:3.22

RUN apk add --no-cache bash

COPY --from=build /build/cerca /usr/bin/cerca

VOLUME /app

RUN mkdir -p /app/data

RUN /usr/bin/cerca write-defaults -config /app/cerca.toml -data-dir /app/data

ENTRYPOINT ["/usr/bin/cerca", "-config", "/app/cerca.toml"]
