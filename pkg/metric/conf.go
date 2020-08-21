package metric

import (
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
	"strings"
)

type TcmMetricConfig struct {
	CustomNamespacePrefix string
	CustomProductName     string
	CustomMetricName      string
	MetricNameType        int32
	CustomQueryDimensions []map[string]string
	InstanceLabelNames    []string
	ConstLabels           map[string]string
	StatTypes             []string
	StatPeriodSeconds     int64
	StatNumSamples        int64
	StatDelaySeconds      int64
	AllInstances          bool
	InstanceFilters       map[string]string
	OnlyIncludeInstances  []string
	ExcludeInstances      []string
}

func NewTcmMetricConfigWithMetricYaml(c config.TencentMetric, meta *TcmMeta) (*TcmMetricConfig, error) {
	conf := &TcmMetricConfig{}

	npitems := strings.Split(c.Namespace, "/")
	conf.CustomNamespacePrefix = npitems[0]

	conf.CustomProductName = npitems[1]
	conf.CustomMetricName = c.MetricReName
	if c.MetricNameType != 0 {
		conf.MetricNameType = c.MetricNameType
	} else {
		conf.MetricNameType = 2
	}

	conf.CustomQueryDimensions = []map[string]string{c.Dimensions}
	conf.InstanceLabelNames = c.Labels
	if len(c.Statistics) != 0 {
		for _, st := range c.Statistics {
			conf.StatTypes = append(conf.StatTypes, strings.ToLower(st))
		}
	} else {
		conf.StatTypes = []string{"last"}
	}
	// 自动获取支持的统计周期
	period, err := meta.GetPeriod(c.PeriodSeconds)
	if err != nil {
		return nil, err
	}
	conf.StatPeriodSeconds = period
	conf.StatNumSamples = (c.RangeSeconds / period) + 1
	// 至少采集4个点的数据
	if conf.StatNumSamples < 4 {
		conf.StatNumSamples = 4
	}
	conf.StatDelaySeconds = c.DelaySeconds

	if len(c.Dimensions) == 0 {
		conf.AllInstances = true
	}

	conf.InstanceFilters = c.Filters
	return conf, nil

}

func NewTcmMetricConfigWithProductYaml(c config.TencentProduct, meta *TcmMeta) (*TcmMetricConfig, error) {
	conf := &TcmMetricConfig{}

	npitems := strings.Split(c.Namespace, "/")
	conf.CustomNamespacePrefix = npitems[0]

	conf.CustomProductName = npitems[1]
	conf.CustomMetricName = ""
	if c.MetricNameType != 0 {
		conf.MetricNameType = c.MetricNameType
	} else {
		conf.MetricNameType = 2
	}

	conf.CustomQueryDimensions = c.CustomQueryDimensions
	conf.InstanceLabelNames = c.ExtraLabels
	if len(c.Statistics) != 0 {
		for _, st := range c.Statistics {
			conf.StatTypes = append(conf.StatTypes, strings.ToLower(st))
		}
	} else {
		conf.StatTypes = []string{"last"}
	}

	period, err := meta.GetPeriod(c.PeriodSeconds)
	if err != nil {
		return nil, err
	}
	conf.StatPeriodSeconds = period
	conf.StatNumSamples = (c.RangeSeconds / period) + 1
	if conf.StatNumSamples < 4 {
		conf.StatNumSamples = 4
	}
	conf.StatDelaySeconds = c.DelaySeconds
	conf.AllInstances = c.AllInstances
	conf.InstanceFilters = c.InstanceFilters
	conf.OnlyIncludeInstances = c.OnlyIncludeInstances
	conf.ExcludeInstances = c.ExcludeInstances

	return conf, nil

}
