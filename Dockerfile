FROM golang:1.8.1-alpine as builder

COPY ./ /go/src/github.com/infinityworks/docker-hub-exporter/

WORKDIR /go/src/github.com/infinityworks/docker-hub-exporter/cmd/exporter/

RUN apk --update add ca-certificates \
    && apk --update add --virtual build-deps git

RUN go get \
 && go test ./... \
 && GOOS=linux go build -o exporter .

FROM alpine

EXPOSE 9170

RUN addgroup exporter \
     && adduser -S -G exporter exporter \
     && apk --update --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/infinityworks/docker-hub-exporter/cmd/exporter/exporter .

USER exporter

ENTRYPOINT ["/exporter"]