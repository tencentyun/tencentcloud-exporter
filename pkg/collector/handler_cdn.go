package collector

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

const (
	CdnNamespace = "QCE/CDN"
)

func init() {
	registerHandler(CdnNamespace, defaultHandlerEnabled, NewCdnHandler)
}

type cdnHandler struct {
	baseProductHandler
}

func (h *cdnHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *cdnHandler) GetNamespace() string {
	return CdnNamespace
}

func (h *cdnHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	return true
}

func (h *cdnHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
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

func (h *cdnHandler) checkMonitorQueryKeys(m *metric.TcmMetric, ql map[string]string) bool {
	for k := range ql {
		if !util.IsStrInList(m.Meta.SupportDimensions, k) {
			level.Error(h.logger).Log("msg", fmt.Sprintf("not found %s in supportQueryDimensions", k),
				"ql", fmt.Sprintf("%v", ql))
			return false
		}
	}
	return true
}

func NewCdnHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &cdnHandler{
		baseProductHandler{
			collector: c,
			logger:    logger,
		},
	}
	return
}
