package collector

import (
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	CbsNamespace     = "QCE/BLOCK_STORAGE"
	CbsInstanceidKey = "diskId"
)

func init() {
	registerHandler(CbsNamespace, defaultHandlerEnabled, NewCbsHandler)
}

type cbsHandler struct {
	baseProductHandler
}

func (h *cbsHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *cbsHandler) GetNamespace() string {
	return CbsNamespace
}

func (h *cbsHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	return true
}

func NewCbsHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &cbsHandler{
		baseProductHandler{
			monitorQueryKey: CbsInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
