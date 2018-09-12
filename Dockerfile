FROM golang:1.11-alpine

RUN apk --update upgrade \
  && apk --no-cache --no-progress add git bash make \
  && rm -rf /var/cache/apk/*

WORKDIR /cwmonitor
COPY . /cwmonitor

ENV GO111MODULE=on

RUN make build

FROM alpine:3.8
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=0 /cwmonitor/cwmonitor /cwmonitor
ENTRYPOINT ["/cwmonitor"]
