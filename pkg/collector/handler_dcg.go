package collector

import (
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	DcgNamespace     = "QCE/DCG"
	DcgInstanceidKey = "directConnectGatewayId"
)

func init() {
	registerHandler(DcgNamespace, defaultHandlerEnabled, NewDcgHandler)
}

type DcgHandler struct {
	baseProductHandler
}

func (h *DcgHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *DcgHandler) GetNamespace() string {
	return DcgNamespace
}

func (h *DcgHandler) IsMetricValid(m *metric.TcmMetric) bool {
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

func NewDcgHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &DcgHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: DcgInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
