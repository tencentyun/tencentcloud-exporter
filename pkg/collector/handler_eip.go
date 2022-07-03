package collector

import (
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	EIPNamespace     = "QCE/LB"
	EIPInstanceidKey = "eip"
)

func init() {
	registerHandler(EIPNamespace, defaultHandlerEnabled, NewEIPHandler)
}

type eipHandler struct {
	baseProductHandler
}

func (h *eipHandler) GetNamespace() string {
	return EIPNamespace
}
func (h *eipHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	// ignore node/shard metric, bug for cloud monitor if filter dim
	if len(m.Meta.SupportDimensions) != 1 {
		return false
	}
	return true
}
func NewEIPHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &eipHandler{
		baseProductHandler{
			monitorQueryKey: EIPInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
