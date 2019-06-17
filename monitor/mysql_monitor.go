package monitor

/*what we call this product in prom*/
const MysqlProductName = "mysql"

/*what they can this product in cloud monitor*/
const mysqlNamespaceInMonitor = "QCE/CDB"

func init() {
	funcGetPrimaryKeys[MysqlProductName] = mysqlGetMonitorData
	PrimaryKeys[MysqlProductName] = "InstanceId"
}

func mysqlGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64,
	instances map[string]map[string]interface{}) (allDataRet map[string]map[int64]float64, errRet error) {

	_ = instances

	return getMonitorDataByPrimarykey(mysqlNamespaceInMonitor,
		instanceIds,
		PrimaryKeys[MysqlProductName],
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
