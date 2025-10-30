FROM golang:alpine AS builder
RUN apk add --no-cache ca-certificates tzdata git curl

WORKDIR /workdir
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

ARG APP_VER="unknown"
ARG GITHASH="unknown"
ENV MISE_INSTALL_PATH="/usr/local/bin/mise"
ENV MISE_TRUSTED_CONFIG_PATHS="/workdir"
ENV MISE_YES=1

# install mise
RUN curl --proto '=https' --tlsv1.2 -sSf https://mise.run | sh
# build
RUN mise x just -- just build

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
