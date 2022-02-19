FROM golang:1.17-alpine3.15 AS build

ENV GOPROXY=https://goproxy.cn,direct

RUN mkdir /caixin-feed
COPY . /caixin-feed
WORKDIR /caixin-feed

RUN go build -o caixin-feed main.go

FROM alpine:3.15
COPY --from=build /caixin-feed/caixin-feed /caixin-feed/caixin-feed

CMD ["/caixin-feed/caixin-feed"]