FROM golang:alpine as builder
RUN apk add --no-cache ca-certificates tzdata make git
WORKDIR /workdir
ENV GOEXPERIMENT=loopvar
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
ARG APP_VER
ARG GITHASH
RUN make build APP_VER="${APP_VER}" GITHASH="${GITHASH}"; rm -rf /root/.cache/

FROM alpine:latest as run
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder --link /workdir/build/bin/* /bin/

USER nobody
EXPOSE 8080/tcp
ENTRYPOINT ["/bin/go-camo"]
