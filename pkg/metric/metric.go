package metric

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
	"strings"
	"time"
)

type TcmMetric struct {
	Id           string
	Meta         *TcmMeta
	Labels       *TcmLabels
	Series       map[string]*TcmSeries
	StatPromDesc map[string]*prometheus.Desc
	Conf         *TcmMetricConfig
}

func (m *TcmMetric) LoadSeries(series []*TcmSeries) error {
	for _, s := range series {
		m.Series[s.Id] = s
	}
	return nil
}

func (m *TcmMetric) GetLatestPromMetrics(repo TcmMetricRepository) (pms []prometheus.Metric, err error) {
	st := time.Now().Unix() - m.Conf.StatNumSamples*m.Conf.StatPeriodSeconds

	samplesList, err := repo.ListSamples(m, st, 0)
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
			values, err := m.Labels.GetValues(samples.Series.QueryLabels, samples.Series.Instance)
			if err != nil {
				return nil, err
			}
			pm := prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				point.Value,
				values...,
			)
			pms = append(pms, pm)
		}
	}

	return
}

func (m TcmMetric) GetSeriesSplitByBatch(batch int) (steps [][]*TcmSeries) {
	var series []*TcmSeries
	for _, s := range m.Series {
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

func NewTcmMetric(meta *TcmMeta, conf *TcmMetricConfig) (*TcmMetric, error) {
	id := fmt.Sprintf("%s-%s", meta.Namespace, meta.MetricName)

	labels, err := NewTcmLabels(meta.SupportDimensions, conf.InstanceLabelNames, conf.ConstLabels)
	if err != nil {
		return nil, err
	}

	statDescs := make(map[string]*prometheus.Desc)
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
	var lnames []string
	for _, name := range labels.Names {
		lnames = append(lnames, util.ToUnderlineLower(name))
	}
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
		desc := prometheus.NewDesc(
			fqName,
			help,
			lnames,
			nil,
		)
		statDescs[strings.ToLower(s)] = desc
	}

	m := &TcmMetric{
		Id:           id,
		Meta:         meta,
		Labels:       labels,
		Series:       map[string]*TcmSeries{},
		StatPromDesc: statDescs,
		Conf:         conf,
	}
	return m, nil
}
