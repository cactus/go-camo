FROM alpine:latest
RUN apk add --no-cache ca-certificates
ARG GITHASH
ARG VERSION
COPY build/bin/* /bin/
EXPOSE 8080/tcp
USER nobody
ENTRYPOINT ["/bin/go-camo"]
