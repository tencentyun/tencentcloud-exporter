package monitor

/*what we call this product in prom*/
const CosProductName = "cos"

/*what they can this product in cloud monitor*/
const cosNamespaceInMonitor = "QCE/COS"

func init() {
	funcGetMultiKeys[CosProductName] = cosMultiGetMonitorData
}

func cosMultiGetMonitorData(dimensions map[string]interface{},
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64) (allDataRet map[int64]float64, errRet error) {

	return getMonitorDataByMultipleKeys(cosNamespaceInMonitor,
		dimensions,
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
