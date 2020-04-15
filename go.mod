module github.com/tencentyun/tencentcloud-exporter

go 1.12

require (
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/common v0.9.1
	github.com/prometheus/tsdb v0.7.1 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.73-0.20190704135516-e86c9d8b05ee+incompatible
	github.com/yangwenmai/ratelimit v0.0.0-20180104140304-44221c2292e1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.5
)
