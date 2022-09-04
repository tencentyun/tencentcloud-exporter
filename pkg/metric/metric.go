package metric

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
)

type SeriesCache struct {
	Series map[string]*TcmSeries // 包含的多个时间线
	// need cache it, because some cases DescribeBaseMetrics/GetMonitorData dims not match
	LabelNames map[string]struct{}
}

func newCache() *SeriesCache {
	return &SeriesCache{
		Series:     make(map[string]*TcmSeries),
		LabelNames: make(map[string]struct{}),
	}
}

type Desc struct {
	FQName string
	Help   string
}

// 代表一个指标, 包含多个时间线
type TcmMetric struct {
	Id           string
	Meta         *TcmMeta   // 指标元数据
	Labels       *TcmLabels // 指标labels
	SeriesCache  *SeriesCache
	StatPromDesc map[string]Desc // 按统计纬度的Desc, max、min、avg、last
	Conf         *TcmMetricConfig
	seriesLock   sync.Mutex
}

func (m *TcmMetric) LoadSeries(series []*TcmSeries) error {
	m.seriesLock.Lock()
	defer m.seriesLock.Unlock()

	newSeriesCache := newCache()

	for _, s := range series {
		newSeriesCache.Series[s.Id] = s
		// add label names
		for key, _ := range s.QueryLabels {
			newSeriesCache.LabelNames[key] = struct{}{}
		}
	}
	m.SeriesCache = newSeriesCache
	return nil
}

func (m *TcmMetric) GetLatestPromMetrics(repo TcmMetricRepository) (pms []prometheus.Metric, err error) {
	var st int64
	et := int64(0)
	now := time.Now().Unix()
	if m.Conf.StatDelaySeconds > 0 {
		st = now - m.Conf.StatPeriodSeconds - m.Conf.StatDelaySeconds
		et = now - m.Conf.StatDelaySeconds
	} else {
		st = now - m.Conf.StatNumSamples*m.Conf.StatPeriodSeconds
	}

	samplesList, err := repo.ListSamples(m, st, et)
	if err != nil {
		return nil, err
	}
	for _, samples := range samplesList {
		for st, desc := range m.StatPromDesc {
			var point *TcmSample
			switch st {
			case "last":
				point, err = samples.GetLatestPoint()
				if err != nil {
					return nil, err
				}
			case "max":
				point, err = samples.GetMaxPoint()
				if err != nil {
					return nil, err
				}
			case "min":
				point, err = samples.GetMinPoint()
				if err != nil {
					return nil, err
				}
			case "avg":
				point, err = samples.GetAvgPoint()
				if err != nil {
					return nil, err
				}
			}
			labels := m.Labels.GetValues(samples.Series.QueryLabels, samples.Series.Instance)
			// add all dimensions from cloud monitor into prom labels
			for _, dim := range point.Dimensions {
				labels[*dim.Name] = *dim.Value
			}
			var names []string
			var values []string
			for k, v := range labels {
				names = append(names, util.ToUnderlineLower(k))
				values = append(values, v)
			}
			newDesc := prometheus.NewDesc(
				desc.FQName,
				desc.Help,
				names,
				nil,
			)
			var pm prometheus.Metric
			if m.Conf.StatDelaySeconds > 0 {
				pm = prometheus.NewMetricWithTimestamp(time.Unix(int64(point.Timestamp), 0), prometheus.MustNewConstMetric(
					newDesc,
					prometheus.GaugeValue,
					point.Value,
					values...,
				))
			} else {
				pm = prometheus.MustNewConstMetric(
					newDesc,
					prometheus.GaugeValue,
					point.Value,
					values...,
				)
			}
			pms = append(pms, pm)
		}
	}

	return
}

func (m *TcmMetric) GetSeriesSplitByBatch(batch int) (steps [][]*TcmSeries) {
	var series []*TcmSeries
	for _, s := range m.SeriesCache.Series {
		series = append(series, s)
	}

	total := len(series)
	for i := 0; i < total/batch+1; i++ {
		s := i * batch
		if s >= total {
			continue
		}
		e := i*batch + batch
		if e >= total {
			e = total
		}
		steps = append(steps, series[s:e])
	}
	return
}

// 创建TcmMetric
func NewTcmMetric(meta *TcmMeta, conf *TcmMetricConfig) (*TcmMetric, error) {
	id := fmt.Sprintf("%s-%s", meta.Namespace, meta.MetricName)
	labels, err := NewTcmLabels(meta.SupportDimensions, conf.InstanceLabelNames, conf.ConstLabels)
	if err != nil {
		return nil, err
	}

	statDescs := make(map[string]Desc)
	statType, err := meta.GetStatType(conf.StatPeriodSeconds)
	if err != nil {
		return nil, err
	}
	help := fmt.Sprintf("Metric from %s.%s unit=%s stat=%s Desc=%s",
		meta.Namespace,
		meta.MetricName,
		*meta.m.Unit,
		statType,
		*meta.m.Meaning.Zh,
	)
	for _, s := range conf.StatTypes {
		var st string
		if s == "last" {
			st = strings.ToLower(statType)
		} else {
			st = strings.ToLower(s)
		}

		// 显示的指标名称
		var mn string
		if conf.CustomMetricName != "" {
			mn = conf.CustomMetricName
		} else {
			mn = meta.MetricName
		}

		// 显示的指标名称格式化
		var vmn string
		if conf.MetricNameType == 1 {
			vmn = util.ToUnderlineLower(mn)
		} else {
			vmn = strings.ToLower(mn)
		}
		fqName := fmt.Sprintf("%s_%s_%s_%s",
			conf.CustomNamespacePrefix,
			conf.CustomProductName,
			vmn,
			st,
		)
		fqName = strings.ToLower(fqName)
		statDescs[strings.ToLower(s)] = Desc{
			FQName: fqName,
			Help:   help,
		}
	}

	m := &TcmMetric{
		Id:          id,
		Meta:        meta,
		Labels:      labels,
		SeriesCache: newCache(),

		StatPromDesc: statDescs,
		Conf:         conf,
	}
	return m, nil
}
