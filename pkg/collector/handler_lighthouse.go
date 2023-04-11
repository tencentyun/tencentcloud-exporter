package collector

import (
	"strings"

	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

const (
	LighthouseNamespace     = "QCE/LIGHTHOUSE"
	LighthouseInstanceIDKey = "InstanceId"
)

func init() {
	registerHandler(LighthouseNamespace, defaultHandlerEnabled, NewLighthouseHandler)
}

type lighthouseHandler struct {
	baseProductHandler
}

func (h *lighthouseHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	if !util.IsStrInList(meta.SupportDimensions, LighthouseInstanceIDKey) {
		meta.SupportDimensions = append(meta.SupportDimensions, LighthouseInstanceIDKey)
	}

	return true
}

func (h *lighthouseHandler) GetNamespace() string {
	return LighthouseNamespace
}

func (h *lighthouseHandler) IsMetricValid(m *metric.TcmMetric) bool {
	if util.IsStrInList(CvmInvalidMetricNames, strings.ToLower(m.Meta.MetricName)) {
		return false
	}
	return true
}

func (h *lighthouseHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
	if m.Conf.StatPeriodSeconds < 60 {
		m.Conf.StatPeriodSeconds = 60
	}
	return h.baseProductHandler.GetSeries(m)
}

func NewLighthouseHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &lighthouseHandler{
		baseProductHandler{
			monitorQueryKey: LighthouseInstanceIDKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
