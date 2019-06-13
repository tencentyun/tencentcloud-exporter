package monitor

/*what we call this product in prom*/
const CdnProductName = "cdn"

/*what they can this product in cloud monitor*/
const cdnNamespaceInMonitor = "QCE/CDN"

func init() {
	funcGetMultiKeys[CdnProductName] = cdnMultiGetMonitorData
}

func cdnMultiGetMonitorData(dimensions map[string]interface{},
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64) (allDataRet map[int64]float64, errRet error) {

	return getMonitorDataByMultipleKeys(cdnNamespaceInMonitor,
		dimensions,
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
