package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	TdmqNamespace     = "QCE/TDMQ"
	TdmqInstanceidKey = "InstanceId"
)

func init() {
	registerHandler(TdmqNamespace, defaultHandlerEnabled, NewTdmqHandler)
	excludeMetricName = map[string]string{
		"LogVolume":           "LogVolume",
		"CurrentBackupVolume": "CurrentBackupVolume",
		"DataVolume":          "DataVolume",
		"FreeBackupVolume":    "FreeBackupVolume",
		"BillingBackupVolume": "BillingBackupVolume",
	}
}

type tdmqHandler struct {
	baseProductHandler
}

func (h *tdmqHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *tdmqHandler) GetNamespace() string {
	return TdmqNamespace
}

func (h *tdmqHandler) IsMetricVaild(m *metric.TcmMetric) bool {
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

func NewTdmqHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &tdmqHandler{
		baseProductHandler{
			monitorQueryKey: TdmqInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
