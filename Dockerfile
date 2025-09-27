FROM golang:1.25.1-alpine3.22 AS build

ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=1

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

RUN mkdir /caixin-feed
COPY . /caixin-feed
WORKDIR /caixin-feed

RUN go build -o caixin-feed main.go

FROM alpine:3.22

# Install SQLite runtime library
RUN apk add --no-cache sqlite

COPY --from=build /caixin-feed/caixin-feed /caixin-feed/caixin-feed

CMD ["/caixin-feed/caixin-feed"]