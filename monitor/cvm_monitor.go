package monitor

/*what we call this product in prom*/
const CvmProductName = "cvm"

/*what they can this product in cloud monitor*/
const cvmNamespaceInMonitor = "QCE/CVM"

func init() {
	funcGetPrimaryKeys[CvmProductName] = cvmGetMonitorData
	PrimaryKeys[CvmProductName] = "InstanceId"
}

func cvmGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64,
	instances map[string]map[string]interface{}) (allDataRet map[string]map[int64]float64, errRet error) {

	return getMonitorDataByPrimarykey(cvmNamespaceInMonitor,
		instanceIds,
		PrimaryKeys[CvmProductName],
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
