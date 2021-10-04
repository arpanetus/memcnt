FROM golang:1.17-alpine as builder

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/github.com/arpanetus/memcnt

COPY . .

ENV CGO_ENABLED=0

RUN go get -d -v && go build -o /bin/memcnt

FROM scratch

WORKDIR /memcnt
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/memcnt /memcnt/memcnt

EXPOSE 8080
ENTRYPOINT ["/memcnt/memcnt"]