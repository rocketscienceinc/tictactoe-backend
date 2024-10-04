FROM golang:1.23-alpine3.19 AS builder

RUN apk --no-cache add git gcc g++
COPY . /srv

ARG PAT
ARG RELEASE

RUN set -x \
    && cd /srv/ \
    && go build -ldflags="-X 'service/cmd.Release=${RELEASE}'" -o /go/bin/app main.go


FROM alpine:3.19
RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/ /usr/local/bin/
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip

CMD ["app"]
