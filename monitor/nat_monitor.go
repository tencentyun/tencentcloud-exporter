package monitor

/*what we call this product in prom*/
const NatProductName = "nat"

/*what they can this product in cloud monitor*/
const natNamespaceInMonitor = "QCE/NAT_GATEWAY"

func init() {
	funcGetPrimaryKeys[NatProductName] = natGetMonitorData
	PrimaryKeys[NatProductName] = "InstanceId"
}

func natGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64,
	instances map[string]map[string]interface{}) (allDataRet map[string]map[int64]float64, errRet error) {

	_ = instances

	return getMonitorDataByPrimarykey(natNamespaceInMonitor,
		instanceIds,
		"natId", //in redis they call nat instance id  "natId" in monitor.
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
