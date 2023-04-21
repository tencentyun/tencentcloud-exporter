package collector

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
	"time"
)

const (
	VbcNamespace     = "QCE/VBC"
	VbcInstanceidKey = "CcnId"
)

var (
	BetweenRegionMetricNames = []string{
		"OutDropBandwidth", "InBandwidthRate", "OutBandwidthRate", "OutDropPkg", "OutDropPkgRate", "InBandwidth", "InPkg", "OutPkg", "OutBandwidth",
	}
	SingleRegionMetricNames = []string{
		"Regioninbandwidthbm", "Regionoutbandwidthbm", "Regionoutdropbandwidthbm", "Regioninpkgbm", "Regionoutbandwidthbmrate", "Regionoutdroppkgbmrate", "Regionoutpkgbm", "Regionoutdroppkgbm",
	}
)
var conf *config.TencentConfig

func init() {
	registerHandler(VbcNamespace, defaultHandlerEnabled, NewVbcHandler)
}

type VbcHandler struct {
	baseProductHandler
	dRegionRepo instance.VbcTcInstanceDRegionRepository
}

func (h *VbcHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *VbcHandler) GetNamespace() string {
	return VbcNamespace
}

func (h *VbcHandler) IsMetricValid(m *metric.TcmMetric) bool {

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

func (h *VbcHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {

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
func (h *VbcHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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
				"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *VbcHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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
				"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *VbcHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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
func (h *VbcHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	if util.IsStrInList(SingleRegionMetricNames, m.Meta.MetricName) {
		return h.getSingleRegionSeries(m, ins)
	} else {
		return h.getBetweenRegionSeries(m, ins)
	}
}

func (h *VbcHandler) getSingleRegionSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	ql := map[string]string{
		h.monitorQueryKey: ins.GetMonitorQueryKey(),
		"SRegion":         h.collector.Conf.Credential.Region,
	}
	s, err := metric.NewTcmSeries(m, ql, ins)
	if err != nil {
		return nil, err
	}
	series = append(series, s)

	return series, nil
}

func (h *VbcHandler) getBetweenRegionSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	dRegionResp, err := h.dRegionRepo.GetVbcDRegionInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, v := range dRegionResp.Response.CcnRegionBandwidthLimitSet {
		ql := map[string]string{
			h.monitorQueryKey: ins.GetMonitorQueryKey(),
			"SRegion":         h.collector.Conf.Credential.Region,
			"DRegion":         *v.Region,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)
	}

	return series, nil
}
func (h *VbcHandler) checkMonitorQueryKeys(m *metric.TcmMetric, ql map[string]string) bool {
	for k := range ql {
		if !util.IsStrInList(m.Meta.SupportDimensions, k) {
			level.Error(h.logger).Log("msg", fmt.Sprintf("not found %s in supportQueryDimensions", k),
				"ql", fmt.Sprintf("%v", ql))
			return false
		}
	}
	return true
}

func NewVbcHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	dRegionRepo, err := instance.NewVbcTcInstanceDRegionRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	relodInterval := time.Duration(c.ProductConf.RelodIntervalMinutes * int64(time.Minute))
	dRegionRepoCahe := instance.NewVbcTcInstanceDRegionRepositoryCache(dRegionRepo, relodInterval, logger)

	handler = &VbcHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: VbcInstanceidKey,
			collector:       c,
			logger:          logger,
		},
		dRegionRepo: dRegionRepoCahe,
	}
	return

}
