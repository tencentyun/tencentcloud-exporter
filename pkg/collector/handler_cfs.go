package collector

import (
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	CfsNamespace     = "QCE/CFS"
	CfsInstanceIdKey = "FileSystemId"
)

func init() {
	registerHandler(CfsNamespace, defaultHandlerEnabled, NewCfsHandler)
}

type CfsHandler struct {
	baseProductHandler
}

func (h *CfsHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *CfsHandler) GetNamespace() string {
	return CfsNamespace
}

func (h *CfsHandler) IsMetricValid(m *metric.TcmMetric) bool {
	return true
}

func (h *CfsHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
	if m.Conf.StatPeriodSeconds < 60 {
		m.Conf.StatPeriodSeconds = 60
	}
	return h.baseProductHandler.GetSeries(m)
}

func NewCfsHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &CfsHandler{
		baseProductHandler{
			monitorQueryKey: CfsInstanceIdKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
