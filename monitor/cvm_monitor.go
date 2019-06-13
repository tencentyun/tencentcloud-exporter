package monitor

/*what we call this product in prom*/
const CvmProductName = "cvm"

/*what they can this product in cloud monitor*/
const cvmNamespaceInMonitor = "QCE/CVM"

func init() {
	funcGetPrimaryKeys[CvmProductName] = cvmGetMonitorData
	//funcGetMultiKeys[CvmProductName] = cvmMultiGetMonitorData
	PrimaryKeys[CvmProductName] = "InstanceId"
}

func cvmGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64) (allDataRet map[string]map[int64]float64, errRet error) {

	return getMonitorDataByPrimarykey(cvmNamespaceInMonitor,
		instanceIds,
		PrimaryKeys[CvmProductName],
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
func cvmMultiGetMonitorData(dimensions map[string]interface{},
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64) (allDataRet map[int64]float64, errRet error) {

	return getMonitorDataByMultipleKeys(cvmNamespaceInMonitor,
		dimensions,
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
