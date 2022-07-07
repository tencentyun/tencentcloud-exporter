package collector

import (
<<<<<<< HEAD
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
=======
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
>>>>>>> master
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
	"time"
)

const (
	NacosNamespace     = "TSE/NACOS"
	NacosInstanceidKey = "NacosInstanceId"
)

func init() {
	registerHandler(NacosNamespace, defaultHandlerEnabled, NewNacosHandler)
}

type NacosHandler struct {
	baseProductHandler
	podRepo       instance.NacosTcInstancePodRepository
	interfaceRepo instance.NacosTcInstanceInterfaceRepository
}

func (h *NacosHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *NacosHandler) GetNamespace() string {
	return NacosNamespace
}

func (h *NacosHandler) IsMetricVaild(m *metric.TcmMetric) bool {
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
func (h *NacosHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	if m.Conf.IsIncludeOnlyInstance() {
		return h.GetSeriesByOnly(m)
	}

<<<<<<< HEAD
func NewNacosHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
=======
	if m.Conf.IsIncludeAllInstance() {
		return h.GetSeriesByAll(m)
	}

	if m.Conf.IsCustomQueryDimensions() {
		return h.GetSeriesByCustom(m)
	}

	return nil, fmt.Errorf("must config all_instances or only_include_instances or custom_query_dimensions")
}

func (h *NacosHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *NacosHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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
				"metric", m.Meta.MetricName, "instacne", ins.GetInstanceId(), "error", err)
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *NacosHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *NacosHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}

	if util.IsStrInList(dimensions, "Interface") {
		return h.getInterfaceSeries(m, ins)
	} else {
		return h.getPodSeries(m, ins)
	}
}

func (h *NacosHandler) getPodSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	podInfoResp, err := h.podRepo.GetNacosPodInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, podInfo := range podInfoResp.Response.Replicas {

		ql := map[string]string{
			"NacosInstanceId": ins.GetMonitorQueryKey(),
			"PodName":         *podInfo.Name,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)

	}

	return series, nil
}

func (h *NacosHandler) getInterfaceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	interfaceInfoResp, err := h.interfaceRepo.GetNacosInterfaceInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	podInfoResp, err := h.podRepo.GetNacosPodInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, podInfo := range podInfoResp.Response.Replicas {
		for _, interfaceInfo := range interfaceInfoResp.Response.Content {
			ql := map[string]string{
				"NacosInstanceId": ins.GetMonitorQueryKey(),
				"PodName":         *podInfo.Name,
				"Interface":       *interfaceInfo.Interface,
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
func NewNacosHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	podRepo, err := instance.NewNacosTcInstancePodRepository(c.Conf, logger)
	if err != nil {
		return nil, err
	}
	relodInterval := time.Duration(c.ProductConf.RelodIntervalMinutes * int64(time.Minute))
	podRepoCahe := instance.NewTcNacosInstancePodCache(podRepo, relodInterval, logger)

	interfaceRepo, err := instance.NewNacosTcInstanceInterfaceRepository(c.Conf, logger)
	if err != nil {
		return nil, err
	}
	interfaceRepoCahe := instance.NewTcNacosInstanceInterfaceCache(interfaceRepo, relodInterval, logger)
>>>>>>> master
	handler = &NacosHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: NacosInstanceidKey,
			collector:       c,
			logger:          logger,
		},
		podRepo:       podRepoCahe,
		interfaceRepo: interfaceRepoCahe,
	}
	return

}
