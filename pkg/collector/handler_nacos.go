package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	NacosNamespace     = "TSE/NACOS"
	NacosInstanceidKey = "InstanceId"
)

func init() {
	registerHandler(NacosNamespace, defaultHandlerEnabled, NewNacosHandler)
}

type NacosHandler struct {
	baseProductHandler
}

func (h *NacosHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *NacosHandler) GetNamespace() string {
	return NacosNamespace
}

func (h *NacosHandler) IsMetricVaild(m *metric.TcmMetric) bool {
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

func NewNacosHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &NacosHandler{
		baseProductHandler{
			monitorQueryKey: NacosInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
