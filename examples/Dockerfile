FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY build/bin/* /bin/
EXPOSE 8080/tcp
USER nobody
ENTRYPOINT ["/bin/go-camo"]
