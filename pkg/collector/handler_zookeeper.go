package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	ZookeeperNamespace     = "TSE/ZOOKEEPER"
	ZookeeperInstanceidKey = "InstanceId"
)

func init() {
	registerHandler(ZookeeperNamespace, defaultHandlerEnabled, NewTdmqHandler)
	excludeMetricName = map[string]string{
		"LogVolume":           "LogVolume",
		"CurrentBackupVolume": "CurrentBackupVolume",
		"DataVolume":          "DataVolume",
		"FreeBackupVolume":    "FreeBackupVolume",
		"BillingBackupVolume": "BillingBackupVolume",
	}
}

type ZookeeperHandler struct {
	baseProductHandler
}

func (h *ZookeeperHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *ZookeeperHandler) GetNamespace() string {
	return TdmqNamespace
}

func (h *ZookeeperHandler) IsMetricVaild(m *metric.TcmMetric) bool {
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

func NewZookeeperHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &tdmqHandler{
		baseProductHandler{
			monitorQueryKey: ZookeeperInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
