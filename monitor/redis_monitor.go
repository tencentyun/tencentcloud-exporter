package monitor

/*what we call this product in prom*/
const RedisProductName = "redis"

/*what they can this product in cloud monitor*/
const redisNamespaceInMonitor = "QCE/REDIS"

func init() {
	funcGetPrimaryKeys[RedisProductName] = redisGetMonitorData
	PrimaryKeys[RedisProductName] = "InstanceId"
}

func redisGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64) (allDataRet map[string]map[int64]float64, errRet error) {

	return getMonitorDataByPrimarykey(redisNamespaceInMonitor,
		instanceIds,
		"redis_uuid", //in redis they call InstanceId  "redis_uuid" in monitor.
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
