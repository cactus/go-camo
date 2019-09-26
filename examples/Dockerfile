FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY build/bin/* /bin/
EXPOSE 80/tcp
ENTRYPOINT ["/bin/go-camo"]
