package collector

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tencentyun/tencentcloud-exporter/config"
	"github.com/tencentyun/tencentcloud-exporter/monitor"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type SpecailMetric struct {
	PromPrefix   string
	ProductName  string
	Labels       []string
	MetricConfig config.TencentMetric
	PromDeses    map[string]*prometheus.Desc /*  prometheus.name-->prometheus.Desc */
	collector    *Collector
	caches       map[string]*MetricCache
}

func NewSpecailMetric(cm config.TencentMetric, collector *Collector) (ret *SpecailMetric, errRet error) {
	ret = &SpecailMetric{}

	ret.MetricConfig = cm
	ret.collector = collector

	/*if tc_namespace is Tencent/CVM,  ProductName is cvm*/
	items := strings.Split(cm.Namespace, `/`)
	ret.ProductName = strings.ToLower(items[len(items)-1])
	ret.Labels = make([]string, 0, len(cm.Dimensions))

	for dName, _ := range cm.Dimensions {
		ret.Labels = append(ret.Labels, dName)
	}
	sort.Strings(ret.Labels)

	/*tc_cvm_cpu_usage_*/
	promPrefix := fmt.Sprintf("%s_%s_%s_", items[0], items[1], ToUnderlineLower(cm.MetricReName))

	ret.PromPrefix = strings.ToLower(promPrefix)

	ret.PromDeses = make(map[string]*prometheus.Desc, len(cm.Statistics))
	ret.caches = make(map[string]*MetricCache)

	for _, v := range cm.Statistics {
		v = strings.ToLower(v)
		/*tc_cvm_cpu_usage_max*/
		promMetric := ret.PromPrefix + v

		docString := fmt.Sprintf("Metric from %s.%s %s", ret.ProductName, cm.MetricName, v)

		labels := make([]string, 0, len(ret.Labels))
		for _, v := range ret.Labels {
			labels = append(labels, ToUnderlineLower(v))
		}

		desc, err := collector.getExistedDesc(promMetric, labels)
		if err != nil {
			errRet = err
			return
		}
		if desc != nil {
			ret.PromDeses[v] = desc
		} else {
			ret.PromDeses[v] = prometheus.NewDesc(promMetric, docString, labels, nil)
			collector.saveDesc(promMetric, labels, ret.PromDeses[v])
		}
	}
	return
}

func (me *SpecailMetric) collect(ch chan<- prometheus.Metric) (errRet error) {

	funcGetMonitorData := monitor.GetMultiKeyFunc(me.ProductName)

	if funcGetMonitorData == nil {
		errRet = fmt.Errorf("error,this product [%s] get self control monitor datas func not support yet", me.ProductName)
		return
	}

	nowts := time.Now().Unix()
	var cacheMetrics = make([]*prometheus.Metric, 0, len(me.MetricConfig.Statistics))
	for _, statistic := range me.MetricConfig.Statistics {
		statistic = strings.ToLower(statistic)
		cacheMetric := me.caches[statistic]

		if cacheMetric != nil &&
			cacheMetric.metric != nil &&
			cacheMetric.insertTime+metricCacheTime > nowts {
			cacheMetrics = append(cacheMetrics, cacheMetric.metric)
		}
	}

	if len(cacheMetrics) > 0 {
		for _, cacheMetric := range cacheMetrics {
			ch <- *cacheMetric
			log.Debugf("metric read from cache,%s", (*cacheMetric).Desc().String())
		}
		return
	}

	datas, err := funcGetMonitorData(me.MetricConfig.Dimensions,
		me.MetricConfig.MetricName,
		me.MetricConfig.PeriodSeconds,
		me.MetricConfig.RangeSeconds,
		me.MetricConfig.DelaySeconds)
	if err != nil {
		me.collector.reportSdkError(monitorProductName)
		return
	}

	nowts = time.Now().Unix()

	for _, statistic := range me.MetricConfig.Statistics {

		statistic = strings.ToLower(statistic)

		funcStatistic := SupportStatistic[statistic]
		if funcStatistic == nil {
			log.Errorf("can not be here, not support statistic [%s ] yet ", statistic)
			continue
		}
		var (
			statisticRet, lastTime, _ = funcStatistic(datas)
			promDesc                  = me.PromDeses[statistic]
			labels                    = me.getLabels()
		)

		proMetric := prometheus.MustNewConstMetric(promDesc, prometheus.GaugeValue, float64(statisticRet), labels...)
		me.caches[statistic] = &MetricCache{
			metric:     &proMetric,
			insertTime: nowts,
		}
		ch <- proMetric
		_ = lastTime
		//ch <- prometheus.NewMetricWithTimestamp(time.Unix(lastTime, 0), proMetric)
	}
	return
}

/*
	Gets a list of tags that the product instance needs to be reported to the prom
*/
func (me *SpecailMetric) getLabels() (ret []string) {

	ret = make([]string, 0, len(me.Labels))
	for _, dimension := range me.Labels {
		value := fmt.Sprintf("%v", me.MetricConfig.Dimensions[dimension])
		ret = append(ret, value)
	}
	return
}
