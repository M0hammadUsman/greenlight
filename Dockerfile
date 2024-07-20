FROM golang:1.22.5-alpine3.20 AS builder
LABEL authors="usman"
WORKDIR /home/usr/app
RUN apk --no-cache add make
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN make build/docker

FROM alpine:3.19
ENV HOME=/home/usr/app/bin
WORKDIR $HOME
COPY --from=builder $HOME/greenlight $HOME
RUN addgroup -S greenlightGroup && adduser -S greenlight -G greenlightGroup
USER greenlight:greenlightGroup
ENTRYPOINT ["./greenlight", "-db-dsn=${GREENLIGHT_DB_DSN}"]

