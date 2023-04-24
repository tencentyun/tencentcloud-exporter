package collector

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

const (
	RedisMemNamespace     = "QCE/REDIS_MEM"
	RedisMemInstanceidKey = "instanceid"
)

var (
//	RedisMemInstanceMetricNames = []string{
//		"cpuutil", "cpumaxutil", "memused", "memutil", "memmaxutil", "keys", "expired", "evicted", "connections",
//		"connectionsutil", "inflow", "inbandwidthutil", "inflowlimit", "outflow", "outbandwidthutil",
//		"outflowlimit", "latencyavg", "latencymax", "latencyread", "latencywrite", "latencyother",
//		"commands", "cmdread", "cmdwrite", "cmdother", "cmdbigvalue", "cmdkeycount", "cmdmget", "cmdslow",
//		"cmdhits", "cmdmiss", "cmderr", "cmdhitsratio",
//	}
//
//	RedisMemProxyMetricNames = []string{
//		"cpuutilproxy", "commandsproxy", "cmdkeycountproxy", "cmdmgetproxy", "cmderrproxy", "cmdbigvalueproxy",
//		"connectionsproxy", "connectionsutilproxy", "inflowproxy", "inbandwidthutilproxy",
//		"inflowlimitproxy", "outflowproxy", "outbandwidthutilproxy", "outflowlimitproxy",
//		"latencyavgproxy", "latencymaxproxy", "latencyreadproxy", "latencywriteproxy", "latencyotherproxy",
//	}
//
//	RedisMemNodeMetricNames = []string{
//		"cpuutilnode", "connectionsnode", "connectionsutilnode", "memusednode", "memutilnode",
//		"keysnode", "expirednode", "evictednode", "repldelaynode", "commandsnode", "cmdreadnode",
//		"cmdwritenode", "cmdothernode", "cmdslownode", "cmdhitsnode", "cmdmissnode", "cmdhitsrationode",
//	}
)

func init() {
	registerHandler(RedisMemNamespace, defaultHandlerEnabled, NewRedisMemHandler)
}

type redisMemHandler struct {
	baseProductHandler

	nodeRepo instance.RedisTcInstanceNodeRepository
}

func (h *redisMemHandler) GetNamespace() string {
	return RedisMemNamespace
}

func (h *redisMemHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *redisMemHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *redisMemHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *redisMemHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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
				"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *redisMemHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	if strings.HasSuffix(m.Meta.MetricName, "Proxy") {
		return h.getProxySeries(m, ins)
	}
	if strings.HasSuffix(m.Meta.MetricName, "Node") {
		return h.getNodeSeries(m, ins)
	}
	return h.getInstanceSeries(m, ins)
}

func (h *redisMemHandler) getInstanceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
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

func (h *redisMemHandler) getProxySeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries

	resp, err := h.nodeRepo.GetNodeInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}

	for _, node := range resp.Response.Proxy {
		ql := map[string]string{
			h.monitorQueryKey: ins.GetMonitorQueryKey(),
			"pnodeid":         *node.NodeId,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)
	}

	return series, nil
}

func (h *redisMemHandler) getNodeSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries

	resp, err := h.nodeRepo.GetNodeInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}

	for _, node := range resp.Response.Redis {
		ql := map[string]string{
			h.monitorQueryKey: ins.GetMonitorQueryKey(),
			"rnodeid":         *node.NodeId,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)
	}

	return series, nil
}

func NewRedisMemHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (ProductHandler, error) {
	nodeRepo, err := instance.NewRedisTcInstanceNodeRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	reloadInterval := time.Duration(c.ProductConf.ReloadIntervalMinutes * int64(time.Minute))
	nodeRepoCache := instance.NewTcRedisInstanceNodeCache(nodeRepo, reloadInterval, logger)

	handler := &redisMemHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: RedisMemInstanceidKey,
			collector:       c,
			logger:          logger,
		},
		nodeRepo: nodeRepoCache,
	}
	return handler, nil
}
