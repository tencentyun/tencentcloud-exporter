package collector

import (
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

const (
	DcxNamespace     = "QCE/DCX"
	DcxInstanceidKey = "directConnectConnId"
)

var (
	DcxInvalidMetricNames = []string{"rxbytes", "txbytes"}
)

func init() {
	registerHandler(DcxNamespace, defaultHandlerEnabled, NewDcxHandler)
}

type dcxHandler struct {
	baseProductHandler
}

func (h *dcxHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	if util.IsStrInList(DcxInvalidMetricNames, strings.ToLower(meta.MetricName)) {
		return false
	}
	return true
}

func (h *dcxHandler) GetNamespace() string {
	return DcxNamespace
}

func (h *dcxHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	return true
}

func NewDcxHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &dcxHandler{
		baseProductHandler{
			monitorQueryKey: DcxInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
