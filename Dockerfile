# This Dockerfile contains multiple targets.
# Use 'docker build --target=<name> .' to build one.
#
# Every target has a BIN_NAME argument that must be provided via --build-arg=BIN_NAME=<name>
# when building.


# ===================================
#
#   Non-release images.
#
# ===================================

# go-discover builds the discover binary (which we don't currently publish
# either).
FROM golang:1.19.2-alpine as go-discover
RUN CGO_ENABLED=0 go install github.com/hashicorp/go-discover/cmd/discover@49f60c093101c9c5f6b04d5b1c80164251a761a6

# dev copies the binary from a local build
# -----------------------------------
# BIN_NAME is a requirement in the hashicorp docker github action
FROM alpine:3.17 AS dev

# NAME and VERSION are the name of the software in releases.hashicorp.com
# and the version to download. Example: NAME=consul VERSION=1.2.3.
ARG BIN_NAME=rollouts-plugin-trafficrouter-consul
ARG VERSION
ARG TARGETARCH
ARG TARGETOS

LABEL name=${BIN_NAME} \
      maintainer="Team Consul Kubernetes <team-consul-kubernetes@hashicorp.com>" \
      vendor="HashiCorp" \
      version=${VERSION} \
      release=${VERSION} \
      summary="rollouts-plugin-trafficrouter-consul is a plugin for Argo Rollouts." \
      description="rollouts-plugin-trafficrouter-consul is a plugin for Argo Rollouts."

# Set ARGs as ENV so that they can be used in ENTRYPOINT/CMD
ENV BIN_NAME=${BIN_NAME}
ENV VERSION=${VERSION}

RUN apk add --no-cache ca-certificates libcap openssl su-exec iputils libc6-compat iptables

# Create a non-root user to run the software.
RUN addgroup ${BIN_NAME} && \
    adduser -S -G ${BIN_NAME} 100

COPY --from=go-discover /go/bin/discover /bin/
COPY pkg/bin/linux_${TARGETARCH}/${BIN_NAME} /bin

USER 100
CMD /bin/${BIN_NAME}


# ===================================
#
#   Release images.
#
# ===================================


# default release image
# -----------------------------------
# This Dockerfile creates a production release image for the project. This
# downloads the release from releases.hashicorp.com and therefore requires that
# the release is published before building the Docker image.
#
# We don't rebuild the software because we want the exact checksums and
# binary signatures to match the software and our builds aren't fully
# reproducible currently.
FROM alpine:3.17 AS release-default

ARG BIN_NAME=rollouts-plugin-trafficrouter-consul
ARG PRODUCT_VERSION

LABEL name=${BIN_NAME} \
      maintainer="Team Consul Kubernetes <team-consul-kubernetes@hashicorp.com>" \
      vendor="HashiCorp" \
      version=${PRODUCT_VERSION} \
      release=${PRODUCT_VERSION} \
      summary="rollouts-plugin-trafficrouter-consul is a plugin for Argo Rollouts" \
      description="rollouts-plugin-trafficrouter-consul is a plugin for Argo Rollouts."

# Set ARGs as ENV so that they can be used in ENTRYPOINT/CMD
ENV BIN_NAME=${BIN_NAME}
ENV VERSION=${PRODUCT_VERSION}

RUN apk add --no-cache ca-certificates libcap openssl su-exec iputils libc6-compat iptables

# TARGETOS and TARGETARCH are set automatically when --platform is provided.
ARG TARGETOS
ARG TARGETARCH

# Create a non-root user to run the software.
RUN addgroup ${BIN_NAME} && \
    adduser -S -G ${BIN_NAME} 100

COPY --from=go-discover /go/bin/discover /bin/
COPY dist/${TARGETOS}/${TARGETARCH}/${BIN_NAME} /bin/

USER 100
CMD /bin/${BIN_NAME}

# ===================================
#
#   Set default target to 'dev'.
#
# ===================================
FROM dev