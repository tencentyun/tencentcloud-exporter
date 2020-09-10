package collector

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
	"strings"
	"sync"
)

// 每个产品的指标采集默认实现, 不同的逻辑通过对应的productHandler实现
type TcProductCollector struct {
	Namespace    string
	MetricRepo   metric.TcmMetricRepository
	InstanceRepo instance.TcInstanceRepository
	MetricMap    map[string]*metric.TcmMetric
	InstanceMap  map[string]instance.TcInstance
	Querys       metric.TcmQuerySet
	Conf         *config.TencentConfig
	handler      productHandler
	logger       log.Logger
	lock         sync.Mutex
}

// 指标纬度配置
func (c *TcProductCollector) loadMetricsByMetricConf() (err error) {
	if len(c.MetricMap) == 0 {
		c.MetricMap = make(map[string]*metric.TcmMetric)
	}

	for _, mconf := range c.Conf.GetMetricConfigs(c.Namespace) {
		meta, err := c.MetricRepo.GetMeta(c.Namespace, mconf.MetricName)
		if err != nil {
			level.Error(c.logger).Log("msg", "not found metric meta", "err", err, "Namespace", c.Namespace, "name", mconf.MetricName)
			continue
		}

		m, ok := c.MetricMap[meta.MetricName]
		if !ok {
			conf, err := metric.NewTcmMetricConfigWithMetricYaml(mconf, meta)
			if err != nil {
				level.Error(c.logger).Log("msg", "parse metric config err", "err", err, "Namespace", c.Namespace, "name", mconf.MetricName)
				continue
			}
			nm, err := metric.NewTcmMetric(meta, conf)
			if err != nil {
				level.Error(c.logger).Log("msg", "create metric err", "err", err, "Namespace", c.Namespace, "name", mconf.MetricName)
				continue
			}
			// 指标过滤
			if !c.handler.IsIncludeMetric(nm) {
				level.Error(c.logger).Log("msg", " Metric not support, skip", "Namespace", c.Namespace, "name", nm.Meta.MetricName)
				continue
			}
			c.MetricMap[meta.MetricName] = nm
			m = nm
		}

		series, err := c.handler.GetSeries(m)
		if err != nil {
			level.Error(c.logger).Log("msg", "create metric series err", "err", err, "Namespace", c.Namespace, "name", mconf.MetricName)
			continue
		}

		err = m.LoadSeries(series)
		if err != nil {
			level.Error(c.logger).Log("msg", "load metric series err", "err", err, "Namespace", c.Namespace, "name", mconf.MetricName)
			continue
		}
	}
	return nil
}

// 产品纬度配置
func (c *TcProductCollector) loadMetricsByProductConf() (err error) {
	if len(c.MetricMap) == 0 {
		c.MetricMap = make(map[string]*metric.TcmMetric)
	}

	for _, pconf := range c.Conf.GetProductConfigs(c.Namespace) {
		var metricNames []string

		if len(pconf.OnlyIncludeMetrics) != 0 {
			// 导出指定指标列表
			for _, mname := range pconf.OnlyIncludeMetrics {
				meta, err := c.MetricRepo.GetMeta(c.Namespace, mname)
				if err != nil {
					level.Warn(c.logger).Log("msg", "not found metric meta", "Namespace", c.Namespace, "name", mname)
				} else {
					metricNames = append(metricNames, meta.MetricName)
				}
			}
		} else {
			// 导出该namespace下的所有指标
			var excludeMetrics []string
			if len(pconf.ExcludeMetrics) != 0 {
				for _, em := range pconf.ExcludeMetrics {
					excludeMetrics = append(excludeMetrics, strings.ToLower(em))
				}
			}
			allMetas, err := c.MetricRepo.ListMetaByNamespace(c.Namespace)
			if err != nil {
				return err
			}

			for _, meta := range allMetas {
				if len(excludeMetrics) != 0 && util.IsStrInList(excludeMetrics, strings.ToLower(meta.MetricName)) {
					continue
				}
				metricNames = append(metricNames, meta.MetricName)
			}
		}

		for _, mname := range metricNames {
			meta, err := c.MetricRepo.GetMeta(c.Namespace, mname)
			if err != nil {
				level.Error(c.logger).Log("msg", "Not found metric meta", "Namespace", c.Namespace, "name", mname)
				continue
			}
			// 指标元数据处理, false=跳过
			if !c.handler.CheckMetricMeta(meta) {
				level.Error(c.logger).Log("msg", " Metric meta check fail, skip", "Namespace", c.Namespace, "name", meta.MetricName)
				continue
			}

			m, ok := c.MetricMap[mname]
			if !ok {
				// 创建TcmMetric模型
				conf, err := metric.NewTcmMetricConfigWithProductYaml(pconf, meta)
				if err != nil {
					level.Error(c.logger).Log("msg", "parse metric config err", "err", err, "Namespace", c.Namespace, "name", mname)
					continue
				}
				nm, err := metric.NewTcmMetric(meta, conf)
				if err != nil {
					level.Error(c.logger).Log("msg", "create metric err", "err", err, "Namespace", c.Namespace, "name", mname)
					continue
				}
				// 指标过滤
				if !c.handler.IsIncludeMetric(nm) {
					level.Error(c.logger).Log("msg", " Metric not support, skip", "Namespace", c.Namespace, "name", nm.Meta.MetricName)
					continue
				}
				c.MetricMap[meta.MetricName] = nm
				m = nm
			}

			// 获取该指标下的所有实例纬度查询或自定义纬度查询
			series, err := c.handler.GetSeries(m)
			if err != nil {
				level.Error(c.logger).Log("msg", "create metric series err", "err", err, "Namespace", c.Namespace, "name", mname)
				continue
			}

			err = m.LoadSeries(series)
			if err != nil {
				level.Error(c.logger).Log("msg", "load metric series err", "err", err, "Namespace", c.Namespace, "name", mname)
				continue
			}
		}

	}
	return nil
}

// 一个query管理一个metric的采集
func (c *TcProductCollector) initQuerys() (err error) {
	var numSeries int
	for _, m := range c.MetricMap {
		q, e := metric.NewTcmQuery(m, c.MetricRepo)
		if e != nil {
			return e
		}
		c.Querys = append(c.Querys, q)
		numSeries += len(q.Metric.Series)
	}
	level.Info(c.logger).Log("msg", "Init all query ok", "Namespace", c.Namespace, "numMetric", len(c.Querys), "numSeries", numSeries)
	return
}

// 执行所有指标的采集
func (c *TcProductCollector) Collect(ch chan<- prometheus.Metric) (err error) {
	wg := sync.WaitGroup{}
	wg.Add(len(c.Querys))
	for _, query := range c.Querys {
		go func(q *metric.TcmQuery) {
			defer wg.Done()
			pms, err := q.GetPromMetrics()
			if err != nil {
				level.Error(c.logger).Log(
					"msg", "Get samples fail",
					"err", err,
					"metric", q.Metric.Id,
				)
			} else {
				for _, pm := range pms {
					ch <- pm
				}
			}

		}(query)
	}
	wg.Wait()

	return
}

// 创建新的TcProductCollector, 每个产品一个
func NewTcProductCollector(namespace string, metricRepo metric.TcmMetricRepository, conf *config.TencentConfig, logger log.Logger) (*TcProductCollector, error) {
	factory, exists := handlerFactoryMap[namespace]
	if !exists {
		return nil, fmt.Errorf("product handler not found, Namespace=%s ", namespace)
	}

	var instanceRepoCache instance.TcInstanceRepository
	if !util.IsStrInList(instance.NotSupportInstances, namespace) {
		// 支持实例自动发现的产品
		instanceRepo, err := instance.NewTcInstanceRepository(namespace, conf, logger)
		if err != nil {
			return nil, err
		}
		// 使用instance缓存
		instanceRepoCache = instance.NewTcInstanceCache(instanceRepo, logger)
	}

	c := &TcProductCollector{
		Namespace:    namespace,
		MetricRepo:   metricRepo,
		InstanceRepo: instanceRepoCache,
		Conf:         conf,
		logger:       logger,
	}

	handler, err := factory(c, logger)
	if err != nil {
		return nil, err
	}
	c.handler = handler

	err = c.loadMetricsByMetricConf()
	if err != nil {
		return nil, err
	}
	err = c.loadMetricsByProductConf()
	if err != nil {
		return nil, err
	}
	err = c.initQuerys()
	if err != nil {
		return nil, err
	}
	return c, nil

}
