package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	ZookeeperNamespace     = "TSE/ZOOKEEPER"
	ZookeeperInstanceidKey = "InstanceId"
)

func init() {
	registerHandler(ZookeeperNamespace, defaultHandlerEnabled, NewZookeeperHandler)
}

type ZookeeperHandler struct {
	baseProductHandler
}

func (h *ZookeeperHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *ZookeeperHandler) GetNamespace() string {
	return ZookeeperNamespace
}

func (h *ZookeeperHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	_, ok := excludeMetricName[m.Meta.MetricName]
	if ok {
		return false
	}
	p, err := m.Meta.GetPeriod(m.Conf.StatPeriodSeconds)
	if err != nil {
		return false
	}
	if p != m.Conf.StatPeriodSeconds {
		return false
	}
	return true
}

func NewZookeeperHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &ZookeeperHandler{
		baseProductHandler{
			monitorQueryKey: ZookeeperInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	level.Warn(c.logger).Log("msg", "NewZookeeperHandler")
	return
}
