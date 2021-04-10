FROM alpine
ADD qcloud_exporter /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/qcloud_exporter"]