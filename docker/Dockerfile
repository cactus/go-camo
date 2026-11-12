
# construct build image
FROM golang:alpine AS builder
RUN apk add --no-cache ca-certificates tzdata git curl

WORKDIR /workdir

# mise setup and tool install
COPY --from=ghcr.io/jdx/mise:latest /usr/local/bin/mise /usr/local/bin/mise
# RUN curl --proto '=https' --tlsv1.2 -sSf https://mise.run | sh

# install build tools
ENV MISE_TRUSTED_CONFIG_PATHS="/workdir"
ENV MISE_YES=1
COPY mise.toml mise.lock ./
RUN mise install --locked

# download go modules
COPY go.mod go.sum ./
RUN mise exec -- go mod download

# build
ARG APP_VER="unknown"
ARG GITHASH="unknown"
COPY . ./
RUN mise exec -- just build

# construct run image
FROM alpine:latest AS run
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder --link /workdir/build/bin/* /bin/

ARG APP_VER="unknown"
ARG GITHASH="unknown"
LABEL git-commit="$GITHASH"
LABEL version="$APP_VER"

USER nobody
EXPOSE 8080/tcp
ENTRYPOINT ["/bin/go-camo"]
