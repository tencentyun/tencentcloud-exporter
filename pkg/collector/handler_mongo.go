package collector

import (
	"fmt"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	mongodb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mongodb/v20190725"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

const (
	MongoNamespace     = "QCE/CMONGO"
	MongoInstanceidKey = "target"
)

var (
	MongoClusterMetrics = []string{
		"inserts", "reads", "updates", "deletes", "counts", "aggregates", "clusterconn", "commands", "connper", "clusterdiskusage",
		"qps", "success", "delay10", "delay50", "delay100", "timeouts",
	}
	MongoReplicaMetrics = []string{
		"replicadiskusage", "slavedelay", "oplogreservedtime",
	}
	MongoNodeMetrics = []string{
		"conn", "cpuusage", "memusage", "qr", "qw", "netin", "netout",
	}
)

func init() {
	registerHandler(MongoNamespace, defaultHandlerEnabled, NewMongoHandler)
}

type mongoHandler struct {
	baseProductHandler
}

func (h *mongoHandler) GetNamespace() string {
	return MongoNamespace
}

func (h *mongoHandler) ModifyMetric(m *metric.TcmMetric) error {
	if m.Meta.MetricName == "Commands" {
		if m.Conf.StatPeriodSeconds == 60 {
			// 该指标不支持60统计
			m.Conf.StatPeriodSeconds = 300
		}
	}
	return nil
}

func (h *mongoHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *mongoHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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
				"metric", m.Meta.MetricName, "instacne", insId)
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *mongoHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *mongoHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	for _, ql := range m.Conf.CustomQueryDimensions {
		v, ok := ql[MongoInstanceidKey]
		if !ok {
			return nil, fmt.Errorf("not found %s in queryDimensions", MongoInstanceidKey)
		}
		ins, err := h.collector.InstanceRepo.Get(v)
		if err != nil {
			return nil, err
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		slist = append(slist, s)
	}
	return slist, nil
}

func (h *mongoHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) (slist []*metric.TcmSeries, err error) {
	if util.IsStrInList(MongoClusterMetrics, strings.ToLower(m.Meta.MetricName)) {
		// 集群纬度
		ql := map[string]string{
			MongoInstanceidKey: ins.GetInstanceId(),
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, fmt.Errorf("create metric series fail, metric=%s, instacne=%s",
				m.Meta.MetricName, ins.GetInstanceId())
		}
		slist = append(slist, s)
	} else if util.IsStrInList(MongoReplicaMetrics, strings.ToLower(m.Meta.MetricName)) {
		// 副本集纬度
		meta, ok := ins.GetMeta().(*mongodb.InstanceDetail)
		if !ok {
			return nil, fmt.Errorf("get instacne raw meta fail, metric=%s, instacne=%s",
				m.Meta.MetricName, ins.GetInstanceId())
		}
		for _, rep := range meta.ReplicaSets {
			// cmgo-6ielucen_0
			ql := map[string]string{
				MongoInstanceidKey: *rep.ReplicaSetId,
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				level.Error(h.logger).Log("msg", "Create metric series fail",
					"metric", m.Meta.MetricName, "instacne", *rep.ReplicaSetId)
			} else {
				slist = append(slist, s)
			}
		}
	} else if util.IsStrInList(MongoNodeMetrics, strings.ToLower(m.Meta.MetricName)) {
		// 节点纬度
		meta, ok := ins.GetMeta().(*mongodb.InstanceDetail)
		if !ok {
			return nil, fmt.Errorf("get instacne raw meta fail, metric=%s, instacne=%s",
				m.Meta.MetricName, ins.GetInstanceId())
		}
		for _, rep := range meta.ReplicaSets {
			// cmgo-6ielucen_0-node-primary
			nprimary := fmt.Sprintf("%s-node-%s", *rep.ReplicaSetId, "primary")
			ql := map[string]string{
				MongoInstanceidKey: nprimary,
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				level.Error(h.logger).Log("msg", "Create metric series fail",
					"metric", m.Meta.MetricName, "instacne", nprimary)
			} else {
				slist = append(slist, s)
			}

			for i := 0; i < int(*rep.SecondaryNum); i++ {
				// cmgo-6ielucen_1-node-slave0
				nslave := fmt.Sprintf("%s-node-slave%d", *rep.ReplicaSetId, i)
				ql := map[string]string{
					MongoInstanceidKey: nslave,
				}
				s, err := metric.NewTcmSeries(m, ql, ins)
				if err != nil {
					level.Error(h.logger).Log("msg", "Create metric series fail",
						"metric", m.Meta.MetricName, "instacne", nslave)
				} else {
					slist = append(slist, s)
				}
			}
		}
	}
	return
}

func NewMongoHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &mongoHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: MongoInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return
}
