package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	VbcNamespace       = "QCE/VBC"
	VbcMonitorQueryKey = "CcnId"
)

func init() {
	registerHandler(VbcNamespace, defaultHandlerEnabled, NewVbcHandler)
}

type VbcHandler struct {
	baseProductHandler
}

func (h *VbcHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *VbcHandler) GetNamespace() string {
	return VbcNamespace
}

func (h *VbcHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	return true
}

func NewVbcHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &VbcHandler{
		baseProductHandler{
			monitorQueryKey: VbcMonitorQueryKey,
			collector:       c,
			logger:          logger,
		},
	}
	return
}
