package collector

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
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

func (h *cosHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *cosHandler) GetNamespace() string {
	return CosNamespace
}

func (h *cosHandler) IsMetricValid(m *metric.TcmMetric) bool {
	// cos大部分指标不支持300以下的统计纬度
	if m.Conf.StatPeriodSeconds < 300 {
		m.Conf.StatPeriodSeconds = 300
	}
	return true
}
func (h *cosHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
	if m.Conf.IsIncludeOnlyInstance() {
		return h.GetSeriesByOnly(m)
	}

	if m.Conf.IsIncludeAllInstance() {
		return h.GetSeriesByAll(m)
	}

	if m.Conf.IsCustomQueryDimensions() {
		return h.GetSeriesByCustom(m)
	}

	return nil, fmt.Errorf("must config all_instances or only_include_instances or custom_query_dimensions")
}

func (h *cosHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	insList, err := h.collector.InstanceRepo.ListByFilters(m.Conf.InstanceFilters)
	if err != nil {
		return nil, err
	}
	for _, ins := range insList {
		if len(m.Conf.ExcludeInstances) != 0 && util.IsStrInList(m.Conf.ExcludeInstances, ins.GetInstanceId()) {
			continue
		}
		bucket, err := ins.GetFieldValueByName("Name")
		if err != nil {
			level.Error(h.logger).Log("msg", "projectId not found")
			continue
		}
		ql := map[string]string{
			"bucket": bucket,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
			continue
		}
		slist = append(slist, s)
	}
	return slist, nil
}

func (h *cosHandler) GetSeriesByCustom(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
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

func NewCosHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &cosHandler{
		baseProductHandler{
			collector: c,
			logger:    logger,
		},
	}
	return
}
