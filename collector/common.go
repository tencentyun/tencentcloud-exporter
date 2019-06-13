package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

type funcStatistic func(map[int64]float64) (float64, int64, bool)

/*for tencent cloud api error */
var sdkErrorReportMetricName = "tencent_cloud_sdk_error"

/*what we can  monitor  product*/
var monitorProductName = "monitor"

/*Data(metric) cache time*/
const metricCacheTime = 30

/*Supports the way statistic are calculated*/
var SupportStatistic = map[string]funcStatistic{
	"min": minFromMap,
	"max": maxFromMap,
	"avg": avgFromMap,
	"sum": sumFromMap,
}

/*Number of concurrent requests*/
var numberOfConcurrent = 10

type MetricCache struct {
	insertTime int64
	metric     *prometheus.Metric
}

func ToUnderlineLower(s string) string {

	var interval byte = 'a' - 'A'

	b := make([]byte, 0, len(s))

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += interval
			if i != 0 {
				b = append(b, '_')
			}
		}
		b = append(b, c)
	}
	return string(b)
}

func minFromMap(input map[int64]float64) (min float64, lastTime int64, has bool) {
	if len(input) == 0 {
		return
	}
	has = true
	first := true
	for ts, v := range input {
		if first {
			min = v
			lastTime = ts
			first = false
		} else {
			if min > v {
				min = v
			}
			if ts > lastTime {
				lastTime = ts
			}
		}
	}
	return
}

func maxFromMap(input map[int64]float64) (max float64, lastTime int64, has bool) {
	if len(input) == 0 {
		return
	}
	has = true
	first := true
	for ts, v := range input {
		if first {
			max = v
			lastTime = ts
			first = false
		} else {
			if max < v {
				max = v
			}
			if ts > lastTime {
				lastTime = ts
			}
		}
	}
	return
}

func avgFromMap(input map[int64]float64) (avg float64, lastTime int64, has bool) {
	if len(input) == 0 {
		return
	}
	has = true
	first := true
	var sum float64 = 0
	for ts, v := range input {
		if first {
			lastTime = ts
			first = false
		}
		sum += v
		if ts > lastTime {
			lastTime = ts
		}
	}
	avg = sum / float64(len(input))
	return
}

func sumFromMap(input map[int64]float64) (sum float64, lastTime int64, has bool) {
	if len(input) == 0 {
		return
	}
	has = true
	first := true
	for ts, v := range input {
		if first {
			lastTime = ts
			first = false
		}
		sum += v
		if ts > lastTime {
			lastTime = ts
		}
	}
	return
}
