package monitor

/*what we call this product in prom*/
const DcProductName = "dc"
const DcxProductName = "dcx"

/*what they can this product in cloud monitor*/
const dcNamespaceInMonitor = "QCE/DC"
const dcxNamespaceInMonitor = "QCE/DCX"

func init() {
	funcGetPrimaryKeys[DcProductName] = dcGetMonitorData
	PrimaryKeys[DcProductName] = "DirectConnectId"

	funcGetPrimaryKeys[DcxProductName] = dcxGetMonitorData
	PrimaryKeys[DcxProductName] = "DirectConnectTunnelId"

}

func dcGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64,
	instances map[string]map[string]interface{}) (allDataRet map[string]map[int64]float64, errRet error) {

	return getMonitorDataByPrimarykey(dcNamespaceInMonitor,
		instanceIds,
		"directConnectId",
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}

func dcxGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64,
	instances map[string]map[string]interface{}) (allDataRet map[string]map[int64]float64, errRet error) {

	return getMonitorDataByPrimarykey(dcxNamespaceInMonitor,
		instanceIds,
		"directConnectConnId",
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
