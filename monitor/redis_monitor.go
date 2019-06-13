package monitor

import (
	"fmt"
)

/*what we call this product in prom*/
const RedisProductName = "redis"
const ClusterRedisProductName = "cluster_redis"

/*what they can this product in cloud monitor*/
const redisNamespaceInMonitor = "QCE/REDIS"

/*https://cloud.tencent.com/document/product/239/20018*/
const (
	REDIS_VERSION_OLD_CLUSTER_REDIS  = "1"
	REDIS_VERSION_MASTER_SLAVE_REDIS = "2"
	REDIS_VERSION_MASTER_SLAVE_CKV   = "3"
	REDIS_VERSION_CLUSTER_CKV        = "4"
	REDIS_VERSION_STANDALONE_REDIS   = "5"
	REDIS_VERSION_CLUSTER_REDIS_V4   = "6"
	REDIS_VERSION_CLUSTER_REDIS      = "7"
)

var clusterVersions = map[string]bool{
	REDIS_VERSION_MASTER_SLAVE_REDIS: true,
	REDIS_VERSION_STANDALONE_REDIS:   true,
	REDIS_VERSION_CLUSTER_REDIS_V4:   true,
	REDIS_VERSION_CLUSTER_REDIS:      true,
}

func init() {
	funcGetPrimaryKeys[RedisProductName] = nonClusterRedisGetMonitorData
	funcGetPrimaryKeys[ClusterRedisProductName] = clusterRedisGetMonitorData
	PrimaryKeys[RedisProductName] = "InstanceId"
}

func nonClusterRedisGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64,
	instances map[string]map[string]interface{}) (allDataRet map[string]map[int64]float64,
	errRet error) {

	noneClusterIds := make([]string, 0, len(instanceIds))
	for id, data := range instances {
		redisVersion := fmt.Sprintf("%v", data["Type"])
		if !clusterVersions[redisVersion] {
			noneClusterIds = append(noneClusterIds, id)
		}
	}
	if len(noneClusterIds) == 0 {
		return
	}
	return getMonitorDataByPrimarykey(redisNamespaceInMonitor,
		noneClusterIds,
		"redis_uuid",
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}

func clusterRedisGetMonitorData(instanceIds []string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64,
	instances map[string]map[string]interface{}) (allDataRet map[string]map[int64]float64,
	errRet error) {

	clusterIds := make([]string, 0, len(instanceIds))
	for id, data := range instances {
		redisVersion := fmt.Sprintf("%v", data["Type"])
		if clusterVersions[redisVersion] {
			clusterIds = append(clusterIds, id)
		}
	}
	if len(clusterIds) == 0 {
		return
	}
	return getMonitorDataByPrimarykey(redisNamespaceInMonitor,
		clusterIds,
		"instanceid",
		metricName,
		periodSeconds,
		rangeSeconds,
		delaySeconds)
}
