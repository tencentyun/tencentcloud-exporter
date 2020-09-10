package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
	"strings"
)

const (
	Clb7Namespace     = "QCE/LOADBALANCE"
	Clb7InstanceidKey = "vip"
)

var (
	Clb7ExcludeMetrics = []string{
		"outpkgratio",
		"intrafficratio",
		"inpkgratio",
		"qpsratio",
		"activeconnratio",
		"newactiveconnratio",
		"outtrafficratio",
	}
)

func init() {
	registerHandler(Clb7Namespace, defaultHandlerEnabled, NewClb7Handler)
}

type clb7Handler struct {
	baseProductHandler
}

func (h *clb7Handler) CheckMetricMeta(meta *metric.TcmMeta) bool {
	meta.SupportDimensions = append(meta.SupportDimensions, "vip")
	return true
}

func (h *clb7Handler) GetNamespace() string {
	return Clb7Namespace
}

func (h *clb7Handler) IsIncludeMetric(m *metric.TcmMetric) bool {
	if util.IsStrInList(Clb7ExcludeMetrics, strings.ToLower(m.Meta.MetricName)) {
		return false
	}
	return true
}

func NewClb7Handler(c *TcProductCollector, logger log.Logger) (handler productHandler, err error) {
	handler = &clb7Handler{
		baseProductHandler{
			monitorQueryKey: Clb7InstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return
}
