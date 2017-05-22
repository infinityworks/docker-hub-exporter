FROM golang:1.8.1-alpine as builder

WORKDIR /go/src/github.com/infinityworksltd/docker-hub-exporter/

COPY ./ /go/src/github.com/infinityworksltd/docker-hub-exporter/

RUN apk --update add ca-certificates \
    && apk --update add --virtual build-deps git

RUN go get \
 && go test ./... \
 && GOOS=linux go build -o app .

FROM alpine

RUN addgroup exporter \
     && adduser -S -G exporter exporter \
     && apk --update --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/infinityworksltd/docker-hub-exporter/app .

USER exporter

ENTRYPOINT ["/app"]