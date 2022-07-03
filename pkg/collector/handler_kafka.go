package collector

import (
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	KafkaNamespace     = "QCE/CKAFKA"
	KafkaInstanceIDKey = "instanceId"
)

func init() {
	registerHandler(KafkaNamespace, defaultHandlerEnabled, NewKafkaHandler)
}

type kafkaHandler struct {
	baseProductHandler
}

func (h *kafkaHandler) GetNamespace() string {
	return MariaDBNamespace
}
func (h *kafkaHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	if len(m.Meta.SupportDimensions) != 1 {
		return false
	}
	if m.Meta.SupportDimensions[0] != KafkaInstanceIDKey {
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
func NewKafkaHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &kafkaHandler{
		baseProductHandler{
			monitorQueryKey: KafkaInstanceIDKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
