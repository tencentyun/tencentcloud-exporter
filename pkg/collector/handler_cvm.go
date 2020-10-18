package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
	"strings"
)

const (
	CvmNamespace     = "QCE/CVM"
	CvmInstanceidKey = "InstanceId"
)

var (
	CvmInvalidMetricNames = []string{"dccpuusage", "dcmemusage"}
)

func init() {
	registerHandler(CvmNamespace, defaultHandlerEnabled, NewCvmHandler)
}

type cvmHandler struct {
	baseProductHandler
}

func (h *cvmHandler) CheckMetricMeta(meta *metric.TcmMeta) bool {
	if !util.IsStrInList(meta.SupportDimensions, CvmInstanceidKey) {
		meta.SupportDimensions = append(meta.SupportDimensions, CvmInstanceidKey)
	}

	return true
}

func (h *cvmHandler) GetNamespace() string {
	return CvmNamespace
}

func (h *cvmHandler) IsIncludeMetric(m *metric.TcmMetric) bool {
	if util.IsStrInList(CvmInvalidMetricNames, strings.ToLower(m.Meta.MetricName)) {
		return false
	}
	return true
}

func (h *cvmHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
	if m.Conf.StatPeriodSeconds < 60 {
		m.Conf.StatPeriodSeconds = 60
	}
	return h.baseProductHandler.GetSeries(m)
}

func NewCvmHandler(c *TcProductCollector, logger log.Logger) (handler productHandler, err error) {
	handler = &cvmHandler{
		baseProductHandler{
			monitorQueryKey: CvmInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
