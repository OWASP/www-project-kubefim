# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.7

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-bookworm@sha256:6ef6e30f0ea5c384f6d111cf856e024e3086bbdcb1779da3f3b3fbba0aea53d2 AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" \
    -o /out/kubefim ./cmd/kubefimd

FROM scratch

ARG VERSION=dev
LABEL org.opencontainers.image.title="OWASP KubeFIM" \
      org.opencontainers.image.description="eBPF-based file integrity and process execution monitoring for Kubernetes" \
      org.opencontainers.image.source="https://github.com/OWASP/www-project-kubefim/tree/main/kubefim-src" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.licenses="Apache-2.0"

COPY --from=build /out/kubefim /kubefim

USER 0:0
ENTRYPOINT ["/kubefim"]
