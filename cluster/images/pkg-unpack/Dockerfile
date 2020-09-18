FROM BASEIMAGE

ARG ARCH
ARG TINI_VERSION

ADD pkg-unpack /usr/local/bin/pkg-unpack

USER 1001
ENTRYPOINT ["pkg-unpack"]
