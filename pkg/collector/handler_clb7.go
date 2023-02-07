package collector

import (
	"strings"

	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
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
		"rsptimeout",
		"setreqavg",
		"setreqmax",
		"setrspavg",
		"setrspmax",
		"settotalreq",
		"vrsptimeout",
		"vsettotalreq",
	}
)

func init() {
	registerHandler(Clb7Namespace, defaultHandlerEnabled, NewClb7Handler)
}

type clb7Handler struct {
	baseProductHandler
}

func (h *clb7Handler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	if !util.IsStrInList(meta.SupportDimensions, Clb7InstanceidKey) {
		meta.SupportDimensions = append(meta.SupportDimensions, Clb7InstanceidKey)
	}

	return true
}

func (h *clb7Handler) GetNamespace() string {
	return Clb7Namespace
}

func (h *clb7Handler) IsMetricVaild(m *metric.TcmMetric) bool {
	if util.IsStrInList(Clb7ExcludeMetrics, strings.ToLower(m.Meta.MetricName)) {
		return false
	}
	return true
}

func NewClb7Handler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &clb7Handler{
		baseProductHandler{
			monitorQueryKey: Clb7InstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return
}
