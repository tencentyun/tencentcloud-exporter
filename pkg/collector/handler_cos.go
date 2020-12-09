package collector

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

const (
	CosNamespace = "QCE/COS"
)

func init() {
	registerHandler(CosNamespace, defaultHandlerEnabled, NewCosHandler)
}

var (
	CosSupportDimensions = []string{"appid", "bucket"}
)

type cosHandler struct {
	baseProductHandler
}

func (h *cosHandler) CheckMetricMeta(meta *metric.TcmMeta) bool {
	return true
}

func (h *cosHandler) GetNamespace() string {
	return CosNamespace
}

func (h *cosHandler) IsIncludeMetric(m *metric.TcmMetric) bool {
	// cos大部分指标不支持300以下的统计纬度
	if m.Conf.StatPeriodSeconds < 300 {
		m.Conf.StatPeriodSeconds = 300
	}
	return true
}

func (h *cosHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
	for _, ql := range m.Conf.CustomQueryDimensions {
		if !h.checkMonitorQueryKeys(m, ql) {
			continue
		}

		s, err := metric.NewTcmSeries(m, ql, nil)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail", "metric", m.Meta.MetricName,
				"ql", fmt.Sprintf("%v", ql))
			continue
		}
		slist = append(slist, s)
	}
	return
}

func (h *cosHandler) checkMonitorQueryKeys(m *metric.TcmMetric, ql map[string]string) bool {
	for k := range ql {
		if !util.IsStrInList(CosSupportDimensions, k) {
			level.Error(h.logger).Log("msg", fmt.Sprintf("not found %s in supportQueryDimensions", k),
				"ql", fmt.Sprintf("%v", ql),
				"sd", fmt.Sprintf("%v", m.Meta.SupportDimensions),
			)
			return false
		}
	}
	return true
}

func NewCosHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &cosHandler{
		baseProductHandler{
			collector: c,
			logger:    logger,
		},
	}
	return
}
