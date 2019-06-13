package monitor

/*what we call this product in prom*/
const ClbProductName = "public_clb"

/*what they can this product in cloud monitor*/
const clbNamespaceInMonitor = "QCE/LB_PUBLIC"

func init() {
	funcGetPrimaryKeys[ClbProductName] = clbGetMonitorData
	//funcGetMultiKeys[ClbProductName] = clbMultiGetMonitorData
	PrimaryKeys[ClbProductName] = "LoadBalancerVip"
}

func clbGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64) (allDataRet map[string]map[int64]float64, errRet error) {

	return getMonitorDataByPrimarykey(clbNamespaceInMonitor,
		instanceIds,
		"vip", //in redis they call LoadBalancerVip  "vip" in monitor.
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}

func clbMultiGetMonitorData(dimensions map[string]interface{},
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64) (allDataRet map[int64]float64, errRet error) {

	return getMonitorDataByMultipleKeys(clbNamespaceInMonitor,
		dimensions,
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
