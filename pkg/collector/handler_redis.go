package collector

import (
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

const (
	RedisNamespace     = "QCE/REDIS"
	RedisInstanceidKey = "instanceid"
)

func init() {
	registerHandler(RedisNamespace, defaultHandlerEnabled, NewRedisHandler)
}

var (
	RedisMetricNames = []string{
		"cpuusmin", "storagemin", "storageusmin", "keysmin", "expiredkeysmin", "evictedkeysmin", "connectionsmin", "connectionsusmin",
		"inflowmin", "inflowusmin", "outflowmin", "outflowusmin", "latencymin", "latencygetmin", "latencysetmin", "latencyothermin",
		"qpsmin", "statgetmin", "statsetmin", "statothermin", "bigvaluemin", "slowquerymin", "statsuccessmin", "statmissedmin",
		"cmderrmin", "cachehitratiomin",
	}
	RedisClusterMetricNames = []string{
		"cpuusmin", "cpumaxusmin", "storagemin", "storageusmin", "storagemaxusmin", "keysmin", "expiredkeysmin", "evictedkeysmin",
		"connectionsmin", "connectionsusmin", "inflowmin", "inflowusmin", "outflowmin", "outflowusmin", "latencymin", "latencygetmin",
		"latencysetmin", "latencyothermin", "qpsmin", "statgetmin", "statsetmin", "statothermin", "bigvaluemin", "slowquerymin",
		"statsuccessmin", "statmissedmin", "cmderrmin", "cachehitratiomin",
	}
)

type redisHandler struct {
	baseProductHandler
}

func (h *redisHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *redisHandler) GetNamespace() string {
	return RedisNamespace
}

func (h *redisHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	if strings.ToLower(m.Conf.CustomProductName) == "cluster_redis" {
		if util.IsStrInList(RedisClusterMetricNames, strings.ToLower(m.Meta.MetricName)) {
			return true
		}
	}
	if strings.ToLower(m.Conf.CustomProductName) == "redis" {
		if util.IsStrInList(RedisMetricNames, strings.ToLower(m.Meta.MetricName)) {
			return true
		}
	}
	return false
}

func NewRedisHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &redisHandler{
		baseProductHandler{
			monitorQueryKey: RedisInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
