FROM golang:1.26-alpine AS builder

RUN apk add --no-cache bash git ca-certificates tzdata make

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
RUN make build VERSION=${VERSION} OUTPUT=/bin/hue-dial

FROM scratch

LABEL org.opencontainers.image.source=https://github.com/rantuma/hue-dial

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /bin/hue-dial /hue-dial

ENV CONFIG_PATH=/data/config.json
VOLUME /data

ENTRYPOINT ["/hue-dial"]
