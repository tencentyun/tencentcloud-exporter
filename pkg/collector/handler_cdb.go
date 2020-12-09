package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	CdbNamespace     = "QCE/CDB"
	CdbInstanceidKey = "InstanceId"
)

func init() {
	registerHandler(CdbNamespace, defaultHandlerEnabled, NewCdbHandler)
}

type cdbHandler struct {
	baseProductHandler
}

func (h *cdbHandler) CheckMetricMeta(meta *metric.TcmMeta) bool {
	return true
}

func (h *cdbHandler) GetNamespace() string {
	return CdbNamespace
}

func (h *cdbHandler) IsIncludeMetric(m *metric.TcmMetric) bool {
	return true
}

func NewCdbHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &cdbHandler{
		baseProductHandler{
			monitorQueryKey: CdbInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
