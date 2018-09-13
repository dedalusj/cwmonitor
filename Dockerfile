FROM alpine:3.8

ARG created
ARG version
ARG revision
ARG build_number
LABEL org.opencontainers.image.created="$created"
LABEL org.opencontainers.image.version="$version"
LABEL org.opencontainers.image.revision="$revision"
LABEL org.opencontainers.image.build_number="$build_number"

RUN apk --no-cache add ca-certificates
WORKDIR /
COPY ./cwmonitor /cwmonitor
ENTRYPOINT ["/cwmonitor"]
