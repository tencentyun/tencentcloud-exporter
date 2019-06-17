package monitor

/*what we call this product in prom*/
const ClbProductName = "public_clb"

/*what they can this product in cloud monitor*/
const clbNamespaceInMonitor = "QCE/LB_PUBLIC"

func init() {
	funcGetPrimaryKeys[ClbProductName] = clbGetMonitorData
	PrimaryKeys[ClbProductName] = "LoadBalancerVip"
}

func clbGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64,
	instances map[string]map[string]interface{}) (allDataRet map[string]map[int64]float64, errRet error) {

	_ = instances

	return getMonitorDataByPrimarykey(clbNamespaceInMonitor,
		instanceIds,
		"vip", //in redis they call LoadBalancerVip  "vip" in monitor.
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
