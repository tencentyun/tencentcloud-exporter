package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	ClbNamespace     = "QCE/LB_PUBLIC"
	ClbInstanceidKey = "vip"
)

func init() {
	registerHandler(ClbNamespace, defaultHandlerEnabled, NewClbHandler)
}

type clbHandler struct {
	baseProductHandler
}

func (h *clbHandler) GetNamespace() string {
	return ClbNamespace
}

func (h *clbHandler) IsIncludeMetric(m *metric.TcmMetric) bool {
	return true
}

func NewClbHandler(c *TcProductCollector, logger log.Logger) (handler productHandler, err error) {
	handler = &clbHandler{
		baseProductHandler{
			monitorQueryKey: ClbInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
