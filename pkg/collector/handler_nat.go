package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	NatNamespace       = "QCE/NAT_GATEWAY"
	NatMonitorQueryKey = "natId"
)

func init() {
	registerHandler(NatNamespace, defaultHandlerEnabled, NewNatHandler)
}

type natHandler struct {
	baseProductHandler
}

func (h *natHandler) CheckMetricMeta(meta *metric.TcmMeta) bool {
	return true
}

func (h *natHandler) GetNamespace() string {
	return NatNamespace
}

func (h *natHandler) IsIncludeMetric(m *metric.TcmMetric) bool {
	return true
}

func NewNatHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &natHandler{
		baseProductHandler{
			monitorQueryKey: NatMonitorQueryKey,
			collector:       c,
			logger:          logger,
		},
	}
	return
}
