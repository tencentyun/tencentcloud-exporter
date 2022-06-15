package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	CynosdbNamespace     = "QCE/CYNOSDB_MYSQL"
	CynosdbInstanceidKey = "InstanceId"
)

func init() {
	registerHandler(CynosdbNamespace, defaultHandlerEnabled, NewCynosdbHandler)
}

type CynosdbHandler struct {
	baseProductHandler
}

func (h *CynosdbHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *CynosdbHandler) GetNamespace() string {
	return CynosdbNamespace
}

func (h *CynosdbHandler) IsMetricVaild(m *metric.TcmMetric) bool {
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

func (h *CynosdbHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
	if m.Conf.StatPeriodSeconds < 60 {
		m.Conf.StatPeriodSeconds = 60
	}
	return h.baseProductHandler.GetSeries(m)
}

func NewCynosdbHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &CynosdbHandler{
		baseProductHandler{
			monitorQueryKey: CynosdbInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
