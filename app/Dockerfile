# Dockerfile — the Clayface Portal image.
#
# Multi-stage, and built on the Ansible control node rather than on the lab VM:
# the VM has no route to a registry, so the finished image is exported with
# `docker save` next to the postgres image and shipped as one tarball. That is
# why the compose file pins `pull_policy: never` — the tarball is the only
# source of images on the host.
#
# The vendor directory travels with the source and the build runs with
# -mod=vendor, so nothing here reaches the network either. See .dockerignore,
# which is careful not to exclude it.

FROM golang:1.27-alpine AS build

WORKDIR /src

# go.mod and go.sum first, then the vendor tree: the dependency layer is the
# expensive one and this way it is only rebuilt when a dependency changes.
COPY go.mod go.sum ./
COPY vendor/ vendor/

COPY cmd/ cmd/
COPY internal/ internal/

# CGO_ENABLED=0 gives a static binary, which is what lets the runtime stage be
# a bare alpine with no libc compatibility to worry about. The go:embed
# directives in internal/seed pull the SQL and the sample documents into the
# binary here, so the runtime image needs no data files.
RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -o /out/portal ./cmd/portal

FROM alpine:3.20

# ca-certificates for the /api/fetch feature, which makes outbound HTTPS
# requests. busybox already provides ping, wget and sh, which is what the
# diagnostics page and the healthcheck need — deliberately so, because a slim
# runtime is what makes the command-injection finding legible: there is almost
# nothing in this image except the application and a shell.
# RUN apk add --no-cache ca-certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# The application runs as an unprivileged user. This matters for the lab rather
# than only for tidiness: it is what bounds the command-injection weakness to
# the container's own filesystem and to what this user can read.
RUN addgroup -S -g 10001 portal \
    && adduser -S -D -H -u 10001 -G portal portal

# /data is where the document store lives and is mounted from a named volume.
# Creating it here, already owned by the application user, is what makes the
# volume usable: Docker initialises an empty named volume with the contents and
# ownership of the image's directory at that path, so the first boot finds a
# writable directory instead of a root-owned one.
RUN mkdir -p /data/documents \
    && chown -R portal:portal /data

COPY --from=build /out/portal /usr/local/bin/portal

USER portal

# Documentation only — the compose file publishes no host port for this
# service. nginx is the only thing on the host network, and it reaches the
# portal over the compose network by service name.
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/portal"]
