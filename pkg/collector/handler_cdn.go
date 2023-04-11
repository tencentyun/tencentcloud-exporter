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
	CdnNamespace     = "QCE/CDN"
	CdnInstanceidKey = "domain"
)

func init() {
	registerHandler(CdnNamespace, defaultHandlerEnabled, NewCdnHandler)
}

type cdnHandler struct {
	baseProductHandler
}

func (h *cdnHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *cdnHandler) GetNamespace() string {
	return CdnNamespace
}

func (h *cdnHandler) IsMetricValid(m *metric.TcmMetric) bool {
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

func (h *cdnHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
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
func (h *cdnHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	for _, insId := range m.Conf.OnlyIncludeInstances {
		ins, err := h.collector.InstanceRepo.Get(insId)
		if err != nil {
			level.Error(h.logger).Log("msg", "Instance not found", "id", insId)
			continue
		}
		projectId, err := ins.GetFieldValueByName("ProjectId")
		domain, err := ins.GetFieldValueByName("Domain")
		if err != nil {
			level.Error(h.logger).Log("msg", "projectId not found")
			continue
		}
		ql := map[string]string{
			"domain":    domain,
			"projectId": projectId,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instacne", insId)
			continue
		}
		slist = append(slist, s)
	}
	return slist, nil
}

func (h *cdnHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	insList, err := h.collector.InstanceRepo.ListByFilters(m.Conf.InstanceFilters)
	if err != nil {
		return nil, err
	}
	for _, ins := range insList {
		if len(m.Conf.ExcludeInstances) != 0 && util.IsStrInList(m.Conf.ExcludeInstances, ins.GetInstanceId()) {
			continue
		}
		projectId, err := ins.GetFieldValueByName("ProjectId")
		domain, err := ins.GetFieldValueByName("Domain")
		if err != nil {
			level.Error(h.logger).Log("msg", "projectId not found")
			continue
		}
		ql := map[string]string{
			"domain":    domain,
			"projectId": projectId,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instacne", ins.GetInstanceId())
			continue
		}
		slist = append(slist, s)
	}
	return slist, nil
}

func (h *cdnHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
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
	return slist, nil
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
