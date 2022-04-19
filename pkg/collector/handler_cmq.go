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
	CMQNamespace     = "QCE/CMQ"
	CMQInstanceIDKey = "queueId"
)

func init() {
	registerHandler(CMQNamespace, defaultHandlerEnabled, NewCMQHandler)
}

type cmqHandler struct {
	baseProductHandler
}

func (h *cmqHandler) GetNamespace() string {
	return CMQNamespace
}

func (h *cmqHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *cmqHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	for _, insId := range m.Conf.OnlyIncludeInstances {
		ins, err := h.collector.InstanceRepo.Get(insId)
		if err != nil {
			level.Error(h.logger).Log("msg", "Instance not found", "id", insId)
			continue
		}
		queueName, err := ins.GetFieldValueByName("QueueName")
		if err != nil {
			level.Error(h.logger).Log("msg", "queue name not found")
			continue
		}
		ql := map[string]string{
			h.monitorQueryKey: ins.GetMonitorQueryKey(),
			"queueName":       queueName, // hack, hardcode 🤮
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

func (h *cmqHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	insList, err := h.collector.InstanceRepo.ListByFilters(m.Conf.InstanceFilters)
	if err != nil {
		return nil, err
	}
	for _, ins := range insList {
		if len(m.Conf.ExcludeInstances) != 0 && util.IsStrInList(m.Conf.ExcludeInstances, ins.GetInstanceId()) {
			continue
		}
		queueName, err := ins.GetFieldValueByName("QueueName")
		if err != nil {
			level.Error(h.logger).Log("msg", "queue name not found")
			continue
		}
		ql := map[string]string{
			h.monitorQueryKey: ins.GetMonitorQueryKey(),
			"queueName":       queueName, // hack, hardcode 🤮
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

func NewCMQHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &cmqHandler{
		baseProductHandler{
			monitorQueryKey: CMQInstanceIDKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
