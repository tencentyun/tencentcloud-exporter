package collector

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

var (
	handlerFactoryMap = make(map[string]func(*TcProductCollector, log.Logger) (productHandler, error))
)

type productHandler interface {
	GetNamespace() string
	IsIncludeMetric(m *metric.TcmMetric) bool
	GetSeries(tcmMetric *metric.TcmMetric) (series []*metric.TcmSeries, err error)
}

func registerHandler(namespace string, isDefaultEnabled bool, factory func(*TcProductCollector, log.Logger) (productHandler, error)) {
	handlerFactoryMap[namespace] = factory
}

type baseProductHandler struct {
	monitorQueryKey string
	collector       *TcProductCollector
	logger          log.Logger
}

func (h *baseProductHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
	if len(m.Conf.OnlyIncludeInstances) != 0 {
		for _, insId := range m.Conf.OnlyIncludeInstances {
			ins, err := h.collector.InstanceRepo.Get(insId)
			if err != nil {
				level.Error(h.logger).Log("msg", "Instance not found", "id", insId)
				continue
			}
			ql := map[string]string{
				h.monitorQueryKey: ins.GetInstanceId(),
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				level.Error(h.logger).Log("msg", "Create metric series fail", "metric", m.Meta.MetricName, "instacne", insId)
				continue
			}
			slist = append(slist, s)
		}
	} else if m.Conf.AllInstances {
		insList, err := h.collector.InstanceRepo.ListByFilters(m.Conf.InstanceFilters)
		if err != nil {
			return nil, err
		}
		for _, ins := range insList {
			ql := map[string]string{
				h.monitorQueryKey: ins.GetInstanceId(),
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				level.Error(h.logger).Log("msg", "Create metric series fail", "metric", m.Meta.MetricName, "instacne", ins.GetInstanceId())
				continue
			}
			slist = append(slist, s)
		}
	} else {
		for _, ql := range m.Conf.CustomQueryDimensions {
			v, ok := ql[h.monitorQueryKey]
			if !ok {
				level.Error(h.logger).Log("msg", fmt.Sprintf("not found %s in queryDimensions", h.monitorQueryKey),
					"ql", fmt.Sprintf("%v", ql))
				continue
			}
			ins, err := h.collector.InstanceRepo.Get(v)
			if err != nil {
				level.Error(h.logger).Log("msg", "Instance not found", "id", v)
				continue
			}

			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				level.Error(h.logger).Log("msg", "Create metric series fail", "metric", m.Meta.MetricName, "instacne", ins.GetInstanceId())
				continue
			}
			slist = append(slist, s)
		}
	}
	return
}
