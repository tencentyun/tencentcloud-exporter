package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	VpnxNamespace     = "QCE/VPNX"
	VpnxInstanceidKey = "vpnConnId"
)

func init() {
	registerHandler(VpnxNamespace, defaultHandlerEnabled, NewVpnxHandler)
}

type VpnxHandler struct {
	baseProductHandler
}

func (h *VpnxHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *VpnxHandler) GetNamespace() string {
	return VpnxNamespace
}

func (h *VpnxHandler) IsMetricVaild(m *metric.TcmMetric) bool {
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

func NewVpnxHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &tdmqHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: VpnxInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
