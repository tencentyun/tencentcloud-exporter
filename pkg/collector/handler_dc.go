package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	DcNamespace     = "QCE/DC"
	DcInstanceidKey = "directConnectId"
)

func init() {
	registerHandler(DcNamespace, defaultHandlerEnabled, NewDcHandler)
}

type dcHandler struct {
	baseProductHandler
}

func (h *dcHandler) CheckMetricMeta(meta *metric.TcmMeta) bool {
	return true
}

func (h *dcHandler) GetNamespace() string {
	return DcNamespace
}

func (h *dcHandler) IsIncludeMetric(m *metric.TcmMetric) bool {
	return true
}

func NewDcHandler(c *TcProductCollector, logger log.Logger) (handler productHandler, err error) {
	handler = &dcHandler{
		baseProductHandler{
			monitorQueryKey: DcInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
