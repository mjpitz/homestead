# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.
FROM golang:alpine3.14 AS builder

WORKDIR /opt/homestead
RUN apk add -U tzdata ca-certificates git build-base

COPY . .

RUN go build -o bin/weather-index-builder ./cmd/weather-index-builder

FROM alpine:3.14

WORKDIR /opt/homestead
RUN apk add -U tzdata ca-certificates

COPY --from=builder /opt/homestead/bin /opt/homestead/bin
