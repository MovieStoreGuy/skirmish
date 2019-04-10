FROM golang:1.11.5-alpine AS Builder

WORKDIR /go/src/github.com/MovieStoreGuy/skirmish
COPY . .

ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux
RUN set -x && \
    apk --no-cache add git && \
    go build -ldflags="-s -w" -o /skirmish

FROM alpine:3.8

LABEL Author='Sean (MovieStoreGuy) Marciniak'

COPY --from=Builder /skirmish /usr/bin/skirmish

RUN apk --no-cache add dumb-init ca-certificates && \
    addgroup -S skirmish && \
    adduser  -S skirmish -G skirmish

USER skirmish:skirmish

WORKDIR /user/skirmish

ENTRYPOINT ["dumb-init", "--", "skirmish"]
