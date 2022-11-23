package collector

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
	"time"
)

const (
	QaapNamespace     = "QCE/QAAP"
	QaapInstanceidKey = "channelId"
)

func init() {
	registerHandler(QaapNamespace, defaultHandlerEnabled, NewQaapHandler)
}

type QaapHandler struct {
	baseProductHandler
	tcpListenersRepo instance.QaapTcInstanceTCPListenersRepository
	udpListenersRepo instance.QaapTcInstanceUDPListenersRepository
	commonQaap       instance.CommonQaapTcInstanceRepository
}

func (h *QaapHandler) IsMetricMetaVaild(meta *metric.TcmMeta) bool {
	return true
}

func (h *QaapHandler) GetNamespace() string {
	return QaapNamespace
}

func (h *QaapHandler) IsMetricVaild(m *metric.TcmMetric) bool {
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

func (h *QaapHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *QaapHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *QaapHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *QaapHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *QaapHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}

	if util.IsStrInList(dimensions, "listenerId") {
		return h.getListenerIdSeries(m, ins)
	} else {
		return h.getInstanceSeries(m, ins)
	}
}

func (h *QaapHandler) getInstanceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	commoninfo, err := h.commonQaap.GetCommonQaapInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	level.Info(h.logger).Log("commoninfo", commoninfo)
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

func (h *QaapHandler) getListenerIdSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries

	tcpListenersInfos, err := h.tcpListenersRepo.GetTCPListenersInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, tcpListenersInfo := range tcpListenersInfos.Response.ListenerSet {
		for _, realServerSet := range tcpListenersInfo.RealServerSet {
			ql := map[string]string{
				h.monitorQueryKey:  ins.GetMonitorQueryKey(),
				"listenerId":       *tcpListenersInfo.ListenerId,
				"originServerInfo": *realServerSet.RealServerIP,
				"protocol":         *tcpListenersInfo.Protocol,
				"listenerName":     *tcpListenersInfo.ListenerName,
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				return nil, err
			}
			series = append(series, s)
		}
	}
	udpListenersInfos, err := h.udpListenersRepo.GetUDPListenersInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, udpListenersInfo := range udpListenersInfos.Response.ListenerSet {
		for _, realServerSet := range udpListenersInfo.RealServerSet {
			ql := map[string]string{
				h.monitorQueryKey:  ins.GetMonitorQueryKey(),
				"listenerId":       *udpListenersInfo.ListenerId,
				"originServerInfo": *realServerSet.RealServerIP,
				"protocol":         *udpListenersInfo.Protocol,
				"listenerName":     *udpListenersInfo.ListenerName,
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

func NewQaapHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	tcpListenersRepo, err := instance.NewQaapTcInstanceTCPListenersRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	relodInterval := time.Duration(c.ProductConf.RelodIntervalMinutes * int64(time.Minute))
	tcpListenersRepoCache := instance.NewTcGaapInstanceeTCPListenersCache(tcpListenersRepo, relodInterval, logger)

	udpListenersRepo, err := instance.NewQaapTcInstanceUDPListenersRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	udpListenersRepoCache := instance.NewTcGaapInstanceeUDPListenersCache(udpListenersRepo, relodInterval, logger)

	commonQaap, err := instance.NewCommonQaapTcInstanceRepository(cred, c.Conf, logger)

	handler = &QaapHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: QaapInstanceidKey,
			collector:       c,
			logger:          logger,
		},
		tcpListenersRepo: tcpListenersRepoCache,
		udpListenersRepo: udpListenersRepoCache,
		commonQaap:       commonQaap,
	}
	return

}
