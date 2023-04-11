package collector

import (
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
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

func (h *dcHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *dcHandler) GetNamespace() string {
	return DcNamespace
}

func (h *dcHandler) IsMetricValid(m *metric.TcmMetric) bool {
	return true
}

func NewDcHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &dcHandler{
		baseProductHandler{
			monitorQueryKey: DcInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
