FROM alpine:3.6 as certificates

RUN apk add -U --no-cache ca-certificates

FROM golang:1.11-alpine3.8 as gobuilder

WORKDIR /go/src/github.com/jonathanbeber/dns-proxy

COPY . .

RUN apk add curl git && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
    dep ensure && \
    CGO_ENABLED=0 go build -o /go/bin/dns-proxy

FROM scratch

COPY --from=gobuilder /go/bin/dns-proxy .

COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/dns-proxy"]
