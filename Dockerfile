FROM golang:1.15.0-alpine3.12

ENV GOPROXY=https://goproxy.cn,direct

RUN mkdir /caixin-feed
COPY . /caixin-feed
WORKDIR /caixin-feed

RUN go build -o caixin-feed main.go

CMD ["./caixin-feed"]