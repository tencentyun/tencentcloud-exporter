package collector

import (
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

const (
	TdmqNamespace     = "QCE/TDMQ"
	TdmqInstanceidKey = "tenant"
)

func init() {
	registerHandler(TdmqNamespace, defaultHandlerEnabled, NewTdmqHandler)
}

type tdmqHandler struct {
	baseProductHandler
	namespaceRepo instance.TdmqTcInstanceRocketMQNameSpacesRepository
	topicRepo     instance.TdmqTcInstanceRocketMQTopicsRepository
}

func (h *tdmqHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *tdmqHandler) GetNamespace() string {
	return TdmqNamespace
}

func (h *tdmqHandler) IsMetricVaild(m *metric.TcmMetric) bool {
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

func (h *tdmqHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *tdmqHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *tdmqHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *tdmqHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *tdmqHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	if util.IsStrInList(dimensions, "environmentId") {
		return h.getNamespaceSeries(m, ins)
	} else {
		return h.getInstanceSeries(m, ins)
	}
}

func (h *tdmqHandler) getInstanceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
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

func (h *tdmqHandler) getNamespaceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	namespacesResp, err := h.namespaceRepo.GetRocketMQNamespacesInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespacesResp.Response.Namespaces {
		topicsResp, err := h.topicRepo.GetRocketMQTopicsInfo(ins.GetInstanceId(), *namespace.NamespaceId)
		if err != nil {
			return nil, err
		}
		for _, topic := range topicsResp.Response.Topics {
			ql := map[string]string{
				"tenantId":      ins.GetMonitorQueryKey(),
				"environmentId": *namespace.NamespaceId,
				"topicName":     *topic.Name,
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				return nil, err
			}
			series = append(series, s)
		}
	}
	return series, nil
}

func NewTdmqHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	namespaceRepo, err := instance.NewTdmqTcInstanceRocketMQNameSpacesRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	relodInterval := time.Duration(c.ProductConf.RelodIntervalMinutes * int64(time.Minute))
	namespaceRepoCahe := instance.NewTcTdmqInstanceNamespaceCache(namespaceRepo, relodInterval, logger)

	topicRepo, err := instance.NewTdmqTcInstanceRocketMQTopicsRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	topicRepoCahe := instance.NewTcTdmqInstanceTopicsCache(topicRepo, relodInterval, logger)

	handler = &tdmqHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: TdmqInstanceidKey,
			collector:       c,
			logger:          logger,
		},
		namespaceRepo: namespaceRepoCahe,
		topicRepo:     topicRepoCahe,
	}
	return

}
