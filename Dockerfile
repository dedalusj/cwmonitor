FROM golang:1.11-alpine

RUN apk --update upgrade \
  && apk --no-cache --no-progress add git bash make \
  && rm -rf /var/cache/apk/*

WORKDIR /cwmonitor
COPY . /cwmonitor

ENV GO111MODULE=on

RUN make build

FROM alpine:3.8
ARG version
ARG git_version
ARG build_time
RUN apk --no-cache add ca-certificates
ENV VERSION $version
ENV GIT_VERSION $git_version
ENV BUILD_TIME $build_time
WORKDIR /
COPY --from=0 /cwmonitor/cwmonitor /cwmonitor
ENTRYPOINT ["/cwmonitor"]
