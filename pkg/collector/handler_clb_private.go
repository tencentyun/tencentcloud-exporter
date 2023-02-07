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
	ClbPrivateNamespace     = "QCE/LB_PRIVATE"
	ClbPrivateInstanceidKey = "vip"
)

var (
	ClbPrivateExcludeMetrics = []string{
		"ConnRatio", "OverloadCurConn", "SnatFail", // clb_snat_vip
		"PvvInpkg", "PvvOutpkg", "PvvConnum", "PvvIntraffic", "PvvNewConn", "PvvOuttraffic", // new_vpcid_proto_vip_vport
		"VvIntraffic", "VvInpkg", "VvNewConn", "VvOutpkg", "VvOuttraffic", "VvConnum", // new_vip_vpcid
	}
)

var (
	LbPrivateSupportDimensions = []string{"vip", "vpcId", "loadBalancerPort", "protocol", "lanIp", "port"}
)

func init() {
	registerHandler(ClbPrivateNamespace, defaultHandlerEnabled, NewClbPrivateHandler)
}

type ClbPrivateHandler struct {
	baseProductHandler
}

func (h *ClbPrivateHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *ClbPrivateHandler) GetNamespace() string {
	return ClbPrivateNamespace
}

func (h *ClbPrivateHandler) IsMetricVaild(m *metric.TcmMetric) bool {
	if util.IsStrInList(ClbPrivateExcludeMetrics, m.Meta.MetricName) {
		return false
	}
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	if len(dimensions) == 0 {
		return false
	}
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
func (h *ClbPrivateHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *ClbPrivateHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *ClbPrivateHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *ClbPrivateHandler) GetSeriesByCustom(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
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
func (h *ClbPrivateHandler) checkMonitorQueryKeys(m *metric.TcmMetric, ql map[string]string) bool {
	for k := range ql {
		if !util.IsStrInList(LbPrivateSupportDimensions, k) {
			level.Error(h.logger).Log("msg", fmt.Sprintf("not found %s in supportQueryDimensions", k),
				"ql", fmt.Sprintf("%v", ql),
				"sd", fmt.Sprintf("%v", m.Meta.SupportDimensions),
			)
			return false
		}
	}
	return true
}

func (h *ClbPrivateHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	return h.getClbPrivateSeries(m, ins)
}

func (h *ClbPrivateHandler) getClbPrivateSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	vpcId, err := ins.GetFieldValueByName("VpcId")
	if err != nil {
		level.Error(h.logger).Log("msg", "ClusterId not found")
	}
	ql := map[string]string{
		h.monitorQueryKey: ins.GetMonitorQueryKey(),
		"vpcId":           vpcId,
	}
	s, err := metric.NewTcmSeries(m, ql, ins)
	if err != nil {
		return nil, err
	}
	series = append(series, s)
	return series, nil
}

func NewClbPrivateHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &ClbPrivateHandler{
		baseProductHandler{
			monitorQueryKey: ClbPrivateInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
