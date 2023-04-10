package collector

import (
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	CdbNamespace     = "QCE/CDB"
	CdbInstanceidKey = "InstanceId"
)

var excludeMetricName map[string]string

func init() {
	registerHandler(CdbNamespace, defaultHandlerEnabled, NewCdbHandler)
	excludeMetricName = map[string]string{
		"LogVolume":           "LogVolume",
		"CurrentBackupVolume": "CurrentBackupVolume",
		"DataVolume":          "DataVolume",
		"FreeBackupVolume":    "FreeBackupVolume",
		"BillingBackupVolume": "BillingBackupVolume",
	}
}

type cdbHandler struct {
	baseProductHandler
}

func (h *cdbHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *cdbHandler) GetNamespace() string {
	return CdbNamespace
}

func (h *cdbHandler) IsMetricValid(m *metric.TcmMetric) bool {
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

func NewCdbHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &cdbHandler{
		baseProductHandler{
			monitorQueryKey: CdbInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
