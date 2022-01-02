# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.

FROM alpine:3.14

WORKDIR /opt/homestead
RUN apk add -U tzdata ca-certificates

ARG TARGETOS
ARG TARGETARCH
COPY dist/homestead_${TARGETOS}_${TARGETARCH}.tar.gz /opt/homestead/homestead.tar.gz
RUN mkdir /opt/homestead/bin && \
    cd /opt/homestead/bin && \
    tar zxf /opt/homestead/homestead.tar.gz && \
    rm /opt/homestead/homestead.tar.gz
