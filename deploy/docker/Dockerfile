# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.
FROM golang:alpine3.14 AS builder

WORKDIR /opt/homestead
RUN apk add -U tzdata ca-certificates sqlite sqlite-dev git build-base

COPY . .

RUN go build --tags "libsqlite3" -o bin/weather-index-builder ./cmd/weather-index-builder

FROM alpine:3.14

WORKDIR /opt/homestead
RUN apk add -U tzdata ca-certificates sqlite sqlite-dev

COPY --from=builder /opt/homestead/bin /opt/homestead/bin
