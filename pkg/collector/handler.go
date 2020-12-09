package collector

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

var (
	handlerFactoryMap = make(map[string]func(*TcProductCollector, log.Logger) (ProductHandler, error))
)

// 每个产品的指标处理逻辑
type ProductHandler interface {
	// 获取云监控指标namespace
	GetNamespace() string
	// 对指标元数据做检验和补充
	CheckMetricMeta(meta *metric.TcmMeta) bool
	// 是否包含该指标, ture=包含, false=不包含
	IsIncludeMetric(m *metric.TcmMetric) bool
	// 获取该指标下符合条件的所有实例, 并生成所有的series
	GetSeries(tcmMetric *metric.TcmMetric) (series []*metric.TcmSeries, err error)
}

// 将对应的产品handler注册到Factory中
func registerHandler(namespace string, isDefaultEnabled bool, factory func(*TcProductCollector, log.Logger) (ProductHandler, error)) {
	handlerFactoryMap[namespace] = factory
}

type baseProductHandler struct {
	monitorQueryKey string
	collector       *TcProductCollector
	logger          log.Logger
}

func (h *baseProductHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries

	if m.Conf.IsIncludeOnlyInstance() {
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
				level.Error(h.logger).Log("msg", "Create metric series fail", "metric", m.Meta.MetricName, "instacne", insId)
				continue
			}
			slist = append(slist, s)
		}
		return slist, nil
	}

	if m.Conf.IsIncludeAllInstance() {
		insList, err := h.collector.InstanceRepo.ListByFilters(m.Conf.InstanceFilters)
		if err != nil {
			return nil, err
		}
		for _, ins := range insList {
			ql := map[string]string{
				h.monitorQueryKey: ins.GetMonitorQueryKey(),
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				level.Error(h.logger).Log("msg", "Create metric series fail", "metric", m.Meta.MetricName, "instacne", ins.GetInstanceId())
				continue
			}
			slist = append(slist, s)
		}
		return slist, nil
	}

	if m.Conf.IsCustomQueryDimensions() {
		for _, ql := range m.Conf.CustomQueryDimensions {
			v, ok := ql[h.monitorQueryKey]
			if !ok {
				level.Error(h.logger).Log("msg", fmt.Sprintf("not found %s in queryDimensions", h.monitorQueryKey),
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
				level.Error(h.logger).Log("msg", "Create metric series fail", "err", err, "metric", m.Meta.MetricName, "instacne", ins.GetInstanceId())
				continue
			}
			slist = append(slist, s)
		}
		return slist, nil
	}

	return nil, fmt.Errorf("must config all_instances or only_include_instances or custom_query_dimensions")
}
