FROM golang:1.15.0

ENV GOPROXY=https://goproxy.cn,direct

RUN mkdir /caixin-feed
COPY . /caixin-feed
WORKDIR /caixin-feed

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o caixin-feed main.go

CMD ["./caixin-feed"]