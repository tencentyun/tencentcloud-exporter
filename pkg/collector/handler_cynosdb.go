package collector

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

const (
	CynosdbNamespace     = "QCE/CYNOSDB_MYSQL"
	CynosdbInstanceidKey = "InstanceId"
)

func init() {
	registerHandler(CynosdbNamespace, defaultHandlerEnabled, NewCynosdbHandler)
}

type CynosdbHandler struct {
	baseProductHandler
}

func (h *CynosdbHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *CynosdbHandler) GetNamespace() string {
	return CynosdbNamespace
}

func (h *CynosdbHandler) IsMetricValid(m *metric.TcmMetric) bool {
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

func (h *CynosdbHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *CynosdbHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	for _, insId := range m.Conf.OnlyIncludeInstances {
		ins, err := h.collector.InstanceRepo.Get(insId)
		if err != nil {
			level.Error(h.logger).Log("msg", "Instance not found", "id", insId)
			continue
		}
		sl, err := h.getSeriesByMetricType(m, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instacne", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *CynosdbHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	insList, err := h.collector.InstanceRepo.ListByFilters(m.Conf.InstanceFilters)
	if err != nil {
		return nil, err
	}
	for _, ins := range insList {
		if len(m.Conf.ExcludeInstances) != 0 && util.IsStrInList(m.Conf.ExcludeInstances, ins.GetInstanceId()) {
			continue
		}
		sl, err := h.getSeriesByMetricType(m, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instacne", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *CynosdbHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	for _, ql := range m.Conf.CustomQueryDimensions {
		v, ok := ql[h.monitorQueryKey]
		if !ok {
			level.Error(h.logger).Log(
				"msg", fmt.Sprintf("not found %s in queryDimensions", h.monitorQueryKey),
				"ql", fmt.Sprintf("%v", ql))
			continue
		}
		ins, err := h.collector.InstanceRepo.Get(v)
		if err != nil {
			level.Error(h.logger).Log("msg", "Instance not found", "err", err, "id", v)
			continue
		}

		sl, err := h.getSeriesByMetricType(m, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instacne", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *CynosdbHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	if util.IsStrInList(dimensions, "ClusterId") {
		return h.getClusterIdSeries(m, ins)
	} else {
		return h.getInstanceSeries(m, ins)
	}
}
func (h *CynosdbHandler) getInstanceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries

	ql := map[string]string{
		h.monitorQueryKey: ins.GetMonitorQueryKey(),
	}
	s, err := metric.NewTcmSeries(m, ql, ins)
	if err != nil {
		return nil, err
	}
	series = append(series, s)

	return series, nil
}

func (h *CynosdbHandler) getClusterIdSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	clusterId, err := ins.GetFieldValueByName("ClusterId")
	if err != nil {
		level.Error(h.logger).Log("msg", "ClusterId not found")
	}
	ql := map[string]string{
		h.monitorQueryKey: ins.GetMonitorQueryKey(),
		"ClusterId":       clusterId,
	}
	s, err := metric.NewTcmSeries(m, ql, ins)
	if err != nil {
		return nil, err
	}
	series = append(series, s)
	return series, nil
}

func NewCynosdbHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &CynosdbHandler{
		baseProductHandler{
			monitorQueryKey: CynosdbInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
