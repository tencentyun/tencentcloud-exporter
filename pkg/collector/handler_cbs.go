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
	CbsNamespace     = "QCE/BLOCK_STORAGE"
	CbsInstanceidKey = "diskId"
)

func init() {
	registerHandler(CbsNamespace, defaultHandlerEnabled, NewCbsHandler)
}

type cbsHandler struct {
	baseProductHandler
	instanceInfosRepo instance.CbsTcInstanceInfosRepository
}

func (h *cbsHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *cbsHandler) GetNamespace() string {
	return CbsNamespace
}

func (h *cbsHandler) IsMetricValid(m *metric.TcmMetric) bool {
	// 暂时过滤nvme盘类指标
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	if util.IsStrInList(dimensions, "vmUuid") {
		return false
	}
	return true
}

func (h *cbsHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *cbsHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	if util.IsStrInList(dimensions, "unInstanceId") {
		for _, insId := range m.Conf.OnlyIncludeInstances {
			cvmIds, err := h.instanceInfosRepo.Get(insId)
			if err != nil || len(cvmIds) == 0 {
				level.Error(h.logger).Log("msg", "Instance not found", "id", insId)
				continue
			}
			sl, err := h.getSeriesByMetricType(m, nil, cvmIds)
			if err != nil {
				level.Error(h.logger).Log("msg", "Create metric series fail",
					"metric", m.Meta.MetricName, "instance", cvmIds)
				continue
			}
			slist = append(slist, sl...)
		}
	} else {
		for _, insId := range m.Conf.OnlyIncludeInstances {
			ins, err := h.collector.InstanceRepo.Get(insId)
			if err != nil {
				level.Error(h.logger).Log("msg", "Instance not found", "id", insId)
				continue
			}

			sl, err := h.getSeriesByMetricType(m, ins, nil)
			if err != nil {
				level.Error(h.logger).Log("msg", "Create metric series fail",
					"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
				continue
			}
			slist = append(slist, sl...)
		}
	}

	return slist, nil
}

func (h *cbsHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	if util.IsStrInList(dimensions, "unInstanceId") {
		sl, err := h.getSeriesByMetricType(m, nil, nil)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instance")
		}
		slist = append(slist, sl...)
	} else {
		insList, err := h.collector.InstanceRepo.ListByFilters(m.Conf.InstanceFilters)
		// cvmIds, err := h.instanceInfosRepo.ListByFilters(m.Conf.InstanceFilters)
		if err != nil {
			return nil, err
		}
		for _, ins := range insList {
			if len(m.Conf.ExcludeInstances) != 0 && util.IsStrInList(m.Conf.ExcludeInstances, ins.GetInstanceId()) {
				continue
			}
			sl, err := h.getSeriesByMetricType(m, ins, nil)
			if err != nil {
				level.Error(h.logger).Log("msg", "Create metric series fail",
					"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
				continue
			}
			slist = append(slist, sl...)
		}
	}

	return slist, nil
}

func (h *cbsHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

		sl, err := h.getSeriesByMetricType(m, ins, nil)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *cbsHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance, ids []string) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	if util.IsStrInList(dimensions, "unInstanceId") {
		return h.getInstanceSeries(m, ins, ids)
	} else {
		return h.getCbsSeries(m, ins)
	}
}

func (h *cbsHandler) getCbsSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
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

func (h *cbsHandler) getInstanceSeries(m *metric.TcmMetric, ins instance.TcInstance, ids []string) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	instanceInfos, err := h.instanceInfosRepo.GetInstanceInfosInfoByFilters(ids)
	if err != nil {
		return nil, err
	}
	for _, instanceInfo := range instanceInfos.Response.InstanceSet {

		ql := map[string]string{
			"InstanceId": *instanceInfo.InstanceId,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)
	}

	return series, nil
}

func NewCbsHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	instanceInfosRepoCahe, err := instance.NewCbsTcInstanceInfosRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}

	handler = &cbsHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: CbsInstanceidKey,
			collector:       c,
			logger:          logger,
		},
		instanceInfosRepo: instanceInfosRepoCahe,
	}
	return

}
