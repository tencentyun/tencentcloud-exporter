package collector

import (
	"fmt"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

var (
	handlerFactoryMap = make(map[string]func(common.CredentialIface, *TcProductCollector, log.Logger) (ProductHandler, error))
)

// 每个产品的指标处理逻辑
type ProductHandler interface {
	// 获取云监控指标namespace
	GetNamespace() string
	// 对指标元数据做检验, true=可用, false=跳过
	IsMetricMetaVaild(meta *metric.TcmMeta) bool
	// 修改指标元数据
	ModifyMetricMeta(meta *metric.TcmMeta) error
	// 对指标做校验, true=可用, false=跳过
	IsMetricVaild(m *metric.TcmMetric) bool
	// 修改指标
	ModifyMetric(m *metric.TcmMetric) error
	// 获取该指标下符合条件的所有实例, 并生成所有的series
	GetSeries(tcmMetric *metric.TcmMetric) (series []*metric.TcmSeries, err error)
}

// 将对应的产品handler注册到Factory中
func registerHandler(namespace string, _ bool, factory func(common.CredentialIface, *TcProductCollector, log.Logger) (ProductHandler, error)) {
	handlerFactoryMap[namespace] = factory
}

type baseProductHandler struct {
	monitorQueryKey string
	collector       *TcProductCollector
	logger          log.Logger
}

func (h *baseProductHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *baseProductHandler) ModifyMetricMeta(meta *metric.TcmMeta) error {
	return nil
}

func (h *baseProductHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	p, err := m.Meta.GetPeriod(m.Conf.StatPeriodSeconds)
	if err != nil {
		return false
	}
	if p != m.Conf.StatPeriodSeconds {
		return false
	}
	return true
}

func (h *baseProductHandler) ModifyMetric(m *metric.TcmMetric) error {
	return nil
}

func (h *baseProductHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *baseProductHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	for _, insId := range m.Conf.OnlyIncludeInstances {
		ins, err := h.collector.InstanceRepo.Get(insId)
		if err != nil {
			level.Error(h.logger).Log("msg", "Instance not found", "id", insId)
			continue
		}
		ql := map[string]string{
			h.monitorQueryKey: ins.GetMonitorQueryKey(),
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

func (h *baseProductHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	insList, err := h.collector.InstanceRepo.ListByFilters(m.Conf.InstanceFilters)
	if err != nil {
		return nil, err
	}
	for _, ins := range insList {
		if len(m.Conf.ExcludeInstances) != 0 && util.IsStrInList(m.Conf.ExcludeInstances, ins.GetInstanceId()) {
			continue
		}
		ql := map[string]string{
			h.monitorQueryKey: ins.GetMonitorQueryKey(),
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

func (h *baseProductHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"err", err, "metric", m.Meta.MetricName, "instacne", ins.GetInstanceId())
			continue
		}
		slist = append(slist, s)
	}
	return slist, nil
}
