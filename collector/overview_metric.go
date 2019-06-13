package collector

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tencentyun/tencentcloud-exporter/config"
	"github.com/tencentyun/tencentcloud-exporter/instances"
	"github.com/tencentyun/tencentcloud-exporter/monitor"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type OverviewMetric struct {
	PromPrefix   string
	ProductName  string
	Labels       []string
	MetricConfig config.TencentMetric
	PromDeses    map[string]*prometheus.Desc /*  prometheus.name-->prometheus.Desc */
	collector    *Collector
	caches       map[string]*MetricCache
}

func NewOverviewMetric(cm config.TencentMetric, collector *Collector) (ret *OverviewMetric, errRet error) {
	ret = &OverviewMetric{}

	ret.MetricConfig = cm
	ret.collector = collector

	/*if tc_namespace is Tencent/CVM,  ProductName is cvm*/
	items := strings.Split(cm.Namespace, `/`)
	ret.ProductName = strings.ToLower(items[len(items)-1])

	/*add primary key to dimensions and  uniq(dimensions) */
	uniqMap := make(map[string]bool)
	primaryKey := monitor.PrimaryKeys[ret.ProductName]
	if primaryKey != "" {
		cm.Labels = append([]string{primaryKey}, cm.Labels...)
	}
	for _, v := range cm.Labels {
		uniqMap[v] = true
	}
	ret.Labels = make([]string, 0, len(uniqMap))
	for _, v := range cm.Labels {
		if uniqMap[v] {
			ret.Labels = append(ret.Labels, v)
			uniqMap[v] = false
		}
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

func (me *OverviewMetric) collect(ch chan<- prometheus.Metric) (errRet error) {

	funcGetInstanceInfos := instances.GetInstanceFunc(me.ProductName)

	if funcGetInstanceInfos == nil {
		errRet = fmt.Errorf("error,this product [%s] get instances func not support yet", me.ProductName)
		return
	}

	funcGetMonitorData := monitor.GetPrimaryKeyFunc(me.ProductName)

	if funcGetMonitorData == nil {
		errRet = fmt.Errorf("error,this product [%s] get monitor datas func not support yet", me.ProductName)
		return
	}

	instances, err := funcGetInstanceInfos(me.MetricConfig.Filters)
	if err != nil {
		me.collector.reportSdkError(me.ProductName)
		log.Errorln(err.Error())
	}
	instanceIds := make([]string, 0, len(instances))
	for instanceId := range instances {
		instanceIds = append(instanceIds, instanceId)
	}
	if len(instanceIds) == 0 {
		return
	}

	var cacheMetrics = make([]*prometheus.Metric, 0, len(me.MetricConfig.Statistics))

	nowts := time.Now().Unix()

	for _, instanceId := range instanceIds {
		for _, statistic := range me.MetricConfig.Statistics {
			statistic = strings.ToLower(statistic)
			cacheMetric := me.caches[instanceId+"#"+statistic]
			if cacheMetric != nil &&
				cacheMetric.metric != nil &&
				cacheMetric.insertTime+metricCacheTime > nowts {
				cacheMetrics = append(cacheMetrics, cacheMetric.metric)
			}
		}
	}

	if len(cacheMetrics) > 0 {
		for _, cacheMetric := range cacheMetrics {
			ch <- *cacheMetric
			log.Debugf("metric read from cache,%s", (*cacheMetric).Desc().String())
		}
		return
	}

	allDataRet, err := funcGetMonitorData(instanceIds,
		me.MetricConfig.MetricName,
		me.MetricConfig.PeriodSeconds,
		me.MetricConfig.RangeSeconds,
		me.MetricConfig.DelaySeconds,
		instances)

	if err != nil {
		me.collector.reportSdkError(monitorProductName)
	}
	/*
		Have to deal with allDataRet whether it's wrong or not.
	*/
	nowts = time.Now().Unix()
	for instanceId, datas := range allDataRet {
		if instances[instanceId] == nil {
			log.Errorf("It was a big api bug, because monitor api return a not exist instance id [%s] ", instanceId)
			me.collector.reportSdkError(monitorProductName)
			continue
		}
		if len(datas) == 0 {
			continue
		}
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
				labels                    = me.getLabels(instances[instanceId])
			)

			proMetric := prometheus.MustNewConstMetric(promDesc, prometheus.GaugeValue, float64(statisticRet), labels...)
			me.caches[instanceId+"#"+statistic] = &MetricCache{
				metric:     &proMetric,
				insertTime: nowts,
			}
			ch <- proMetric
			_ = lastTime
			//ch <- prometheus.NewMetricWithTimestamp(time.Unix(lastTime, 0), proMetric)
		}
	}
	return
}

/*
	Gets a list of tags that the product instance needs to be reported to the prom
*/
func (me *OverviewMetric) getLabels(instanceInfo map[string]interface{}) (ret []string) {
	lowerKeyinstanceInfo := make(map[string]interface{}, len(instanceInfo))
	for key, value := range instanceInfo {
		lowerKeyinstanceInfo[strings.ToLower(key)] = value
	}
	ret = make([]string, 0, len(me.Labels))

	for _, dimension := range me.Labels {
		dimension = strings.ToLower(dimension)
		dimensionValue := lowerKeyinstanceInfo[dimension]
		if dimensionValue == nil {
			ret = append(ret, "")
		} else {
			ret = append(ret, fmt.Sprintf("%v", dimensionValue))
		}
	}
	return
}
