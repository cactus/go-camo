FROM golang:1.18-alpine

WORKDIR /app
ADD . /app

RUN apk add --no-cache make git
RUN make build
RUN apk add --no-cache ca-certificates
RUN ls build
RUN cp build/bin/* /bin/

USER 0
ENTRYPOINT []
CMPD []
