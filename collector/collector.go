package collector

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/tencentyun/tencentcloud-exporter/config"
	"github.com/tencentyun/tencentcloud-exporter/instances"
	"github.com/tencentyun/tencentcloud-exporter/monitor"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type MetricReuseDesc struct {
	metricName   string
	metricLabels []string
	desc         *prometheus.Desc
}

type Collector struct {
	locker sync.Mutex
	/*Looking at indicator data from an overview perspective*/
	overviewMetrics []*OverviewMetric

	/*Look at index data from multidimensional perspective*/
	specialDimensionMetrics []*SpecailMetric

	allMetricConfig []config.TencentMetric

	/*Reuse desc and detect duplicates*/
	allPromDesc map[string]*MetricReuseDesc

	sdkErrorDesc *prometheus.Desc
	sdkErrors    map[string]int64
}

func NewCollector(allMetricConfig []config.TencentMetric) (ret *Collector, errRet error) {
	ret = &Collector{}

	ret.overviewMetrics = make([]*OverviewMetric, 0, len(allMetricConfig))
	ret.specialDimensionMetrics = make([]*SpecailMetric, 0, len(allMetricConfig))
	ret.allPromDesc = make(map[string]*MetricReuseDesc)

	ret.allMetricConfig = allMetricConfig

	if errRet = ret.checkConfig(); errRet != nil {
		return
	}
	ret.sdkErrorDesc = prometheus.NewDesc(sdkErrorReportMetricName, "Error fom sdk error", []string{"product"}, nil)
	ret.sdkErrors = make(map[string]int64)
	for _, metricConfig := range allMetricConfig {

		if len(metricConfig.Dimensions) == 0 {
			metric, err := NewOverviewMetric(metricConfig, ret)
			if err != nil {
				errRet = err
				return
			}

			ret.overviewMetrics = append(ret.overviewMetrics, metric)
		} else {
			metric, err := NewSpecailMetric(metricConfig, ret)
			if err != nil {
				errRet = err
				return
			}

			ret.specialDimensionMetrics = append(ret.specialDimensionMetrics, metric)
		}

	}

	return
}

func (me *Collector) getExistedDesc(promMetric string, labels []string) (desc *prometheus.Desc, errRet error) {
	temp := labels[:]
	sort.Strings(temp)
	saved := me.allPromDesc[promMetric]
	if saved == nil {
		return
	}
	if strings.Join(saved.metricLabels, "#") == strings.Join(temp, "#") {
		log.Debugf("prometheus metric desc reuse %s with labels %+v", promMetric, temp)
		desc = saved.desc
		return
	} else {
		errRet = fmt.Errorf("prometheus metric name %s with labels %+v conflict labels %+v", promMetric, saved.metricLabels, temp)
		return
	}
}

func (me *Collector) saveDesc(promMetric string, labels []string, desc *prometheus.Desc) {
	temp := labels[:]
	sort.Strings(temp)

	var reuseDesc MetricReuseDesc
	reuseDesc.desc = desc
	reuseDesc.metricLabels = temp
	reuseDesc.metricName = promMetric

	me.allPromDesc[promMetric] = &reuseDesc
	return
}

/*
	Check that the configurations in the file "yml" are legal
*/
func (me *Collector) checkConfig() (errRet error) {
	for _, metricConfig := range me.allMetricConfig {
		for _, statistic := range metricConfig.Statistics {
			if SupportStatistic[strings.ToLower(statistic)] == nil {
				errRet = fmt.Errorf("not support statistic [%s] yet", statistic)
				return
			}
		}
		items := strings.Split(metricConfig.Namespace, `/`)
		productName := items[len(items)-1]
		productName = strings.ToLower(productName)

		var (
			multiKeyFunc   = monitor.GetMultiKeyFunc(productName)
			primaryKeyFunc = monitor.GetPrimaryKeyFunc(productName)
			instanceFunc   = instances.GetInstanceFunc(productName)
		)

		if multiKeyFunc == nil && primaryKeyFunc == nil {
			errRet = fmt.Errorf("not support product [%s]  yet, need monitor api code.", productName)
			return
		}

		if len(metricConfig.Dimensions) == 0 {
			if instanceFunc == nil || primaryKeyFunc == nil {
				errRet = fmt.Errorf("product [%s] not support [tc_labels,tc_filters] yet, you can use [tc_myself_dimensions]", productName)
				return
			}
		} else {
			if multiKeyFunc == nil {
				errRet = fmt.Errorf("product [%s] not support [tc_myself_dimensions] yet, you can use [tc_labels,tc_filters]", productName)
				return
			}
		}
	}
	return

}

func (me *Collector) Describe(ch chan<- *prometheus.Desc) {
	me.locker.Lock()
	defer me.locker.Unlock()
	for _, v := range me.overviewMetrics {
		for _, desc := range v.PromDeses {
			ch <- desc
		}
	}
	for _, v := range me.specialDimensionMetrics {
		for _, desc := range v.PromDeses {
			ch <- desc
		}
	}

	return
}

func (me *Collector) Collect(ch chan<- prometheus.Metric) {
	me.locker.Lock()
	defer me.locker.Unlock()

	var group sync.WaitGroup

	for _, v := range me.overviewMetrics {
		group.Add(1)
		vTemp := v
		go func() {
			defer group.Done()
			if err := vTemp.collect(ch); err != nil {
				log.Errorf("collect monitor data fail , reason  %s ", err.Error())
			}
		}()
	}

	for _, v := range me.specialDimensionMetrics {
		group.Add(1)
		vTemp := v
		go func() {
			defer group.Done()
			if err := vTemp.collect(ch); err != nil {
				log.Errorf("collect monitor data fail , reason  %s ", err.Error())
			}
		}()
	}

	group.Wait()

	for productName, value := range me.sdkErrors {
		ch <- prometheus.MustNewConstMetric(me.sdkErrorDesc, prometheus.GaugeValue, float64(value), productName)
	}
	me.sdkErrors = make(map[string]int64)
	return
}

/*
	Error reporting for  tencent-sdk-go  API
*/
func (me *Collector) reportSdkError(productName string) {
	me.sdkErrors[productName]++
}
