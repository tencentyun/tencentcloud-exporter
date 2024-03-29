FROM golang:alpine as build-env

# Copy source + vendor
COPY . /go/src/github.com/tencentyun/qcloud-exporter
WORKDIR /go/src/github.com/tencentyun/qcloud-exporter

# Build
ENV GOPATH=/go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -v -a -ldflags "-s -w" -o /go/bin/qcloud_exporter ./cmd/qcloud-exporter/

FROM alpine
COPY --from=build-env /go/bin/qcloud_exporter /usr/bin/qcloud_exporter
RUN apk update
#RUN apk add git
RUN apk add curl
RUN apk add tcpdump
ENTRYPOINT ["qcloud_exporter"]
