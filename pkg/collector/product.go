package collector

import (
	"context"
	"fmt"
	"github.com/tencentyun/tencentcloud-exporter/pkg/constant"
	"strings"
	"sync"
	"time"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
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
	ProductConf  *config.TencentProduct
	handler      ProductHandler
	logger       log.Logger
	lock         sync.RWMutex
}

// 指标纬度配置
func (c *TcProductCollector) LoadMetricsByMetricConf() error {
	if len(c.MetricMap) == 0 {
		c.MetricMap = make(map[string]*metric.TcmMetric)
	}

	for _, mconf := range c.Conf.GetMetricConfigs(c.Namespace) {
		nm, err := c.createMetricWithMetricConf(mconf)
		if err != nil {
			level.Warn(c.logger).Log("msg", "Create metric fail", "err", err,
				"Namespace", c.Namespace, "name", mconf.MetricName)
			continue
		}
		c.MetricMap[nm.Meta.MetricName] = nm

		series, err := c.handler.GetSeries(nm)
		if err != nil {
			level.Error(c.logger).Log("msg", "create metric series err", "err", err,
				"Namespace", c.Namespace, "name", mconf.MetricName)
			continue
		}

		err = nm.LoadSeries(series)
		if err != nil {
			level.Error(c.logger).Log("msg", "load metric series err", "err", err,
				"Namespace", c.Namespace, "name", mconf.MetricName)
			continue
		}
	}
	return nil
}

// 产品纬度配置
func (c *TcProductCollector) LoadMetricsByProductConf() error {
	if len(c.MetricMap) == 0 {
		c.MetricMap = make(map[string]*metric.TcmMetric)
	}

	pconf, err := c.Conf.GetProductConfig(c.Namespace)
	if err != nil {
		return err
	}
	metricNames, err := c.getMetricNames(pconf)
	if err != nil {
		return err
	}
	if c.Namespace == "QCE/QAAP" {
		wg := &sync.WaitGroup{}
		for _, mname := range metricNames {
			wg.Add(1)
			go func(mname string, group *sync.WaitGroup) {
				start := time.Now()
				nm, err := c.createMetricWithProductConf(mname, pconf)
				if err != nil {
					level.Warn(c.logger).Log("msg", "Create metric fail", "err", err, "Namespace", c.Namespace, "name", mname)
					// continue
				}
				if nm == nil {
					// maybe some metric not support
					// continue
				}
				c.lock.Lock()
				c.MetricMap[nm.Meta.MetricName] = nm
				c.lock.Unlock()
				// 获取该指标下的所有实例纬度查询或自定义纬度查询
				series, err := c.handler.GetSeries(nm)
				if err != nil {
					level.Error(c.logger).Log("msg", "create metric series err", "err", err, "Namespace", c.Namespace, "name", mname)
					// continue
				}
				level.Info(c.logger).Log("msg", "found instances", "count", len(series), "Namespace", c.Namespace, "name", mname, "cost", time.Since(start).Milliseconds())
				err = nm.LoadSeries(series)
				if err != nil {
					level.Error(c.logger).Log("msg", "load metric series err", "err", err, "Namespace", c.Namespace, "name", mname)
					// continue
				}
				group.Done()
			}(mname, wg)
		}
		wg.Wait()
	} else {
		for _, mname := range metricNames {
			start := time.Now()
			nm, err := c.createMetricWithProductConf(mname, pconf)
			if err != nil {
				level.Warn(c.logger).Log("msg", "Create metric fail", "err", err, "Namespace", c.Namespace, "name", mname)
				continue
			}
			if nm == nil {
				// maybe some metric not support
				continue
			}
			c.MetricMap[nm.Meta.MetricName] = nm

			// 获取该指标下的所有实例纬度查询或自定义纬度查询
			series, err := c.handler.GetSeries(nm)
			if err != nil {
				level.Error(c.logger).Log("msg", "create metric series err", "err", err, "Namespace", c.Namespace, "name", mname)
				continue
			}
			level.Info(c.logger).Log("msg", "found instances", "count", len(series), "Namespace", c.Namespace, "name", mname, "cost", time.Since(start).Milliseconds())
			err = nm.LoadSeries(series)
			if err != nil {
				level.Error(c.logger).Log("msg", "load metric series err", "err", err, "Namespace", c.Namespace, "name", mname)
				continue
			}
		}
	}
	return nil
}

func (c *TcProductCollector) getMetricNames(pconf config.TencentProduct) ([]string, error) {
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
			return nil, err
		}

		for _, meta := range allMetas {
			if len(excludeMetrics) != 0 && util.IsStrInList(excludeMetrics, strings.ToLower(meta.MetricName)) {
				continue
			}
			metricNames = append(metricNames, meta.MetricName)
		}
	}
	return metricNames, nil
}

func (c *TcProductCollector) createMetricWithProductConf(mname string, pconf config.TencentProduct) (*metric.TcmMetric, error) {
	meta, err := c.MetricRepo.GetMeta(c.Namespace, mname)
	if err != nil {
		return nil, err
	}
	// 指标元数据处理, false=跳过
	if !c.handler.IsMetricMetaValid(meta) {
		return nil, fmt.Errorf("metric not support")
	}
	err = c.handler.ModifyMetricMeta(meta)
	if err != nil {
		return nil, err
	}
	c.lock.RLock()
	m, exists := c.MetricMap[mname]
	c.lock.RUnlock()
	if !exists {
		// 创建TcmMetric模型
		conf, err := metric.NewTcmMetricConfigWithProductYaml(pconf, meta)
		if err != nil {
			return nil, err
		}
		nm, err := metric.NewTcmMetric(meta, conf)
		if err != nil {
			return nil, err
		}
		// 指标过滤
		if !c.handler.IsMetricValid(nm) {
			// ignore invalid metric
			return nil, nil
		}
		err = c.handler.ModifyMetric(nm)
		if err != nil {
			return nil, err
		}
		return nm, nil
	}
	return m, nil
}

func (c *TcProductCollector) createMetricWithMetricConf(mconf config.TencentMetric) (*metric.TcmMetric, error) {
	meta, err := c.MetricRepo.GetMeta(c.Namespace, mconf.MetricName)
	if err != nil {
		return nil, err
	}
	// 指标元数据处理, false=跳过
	if !c.handler.IsMetricMetaValid(meta) {
		return nil, fmt.Errorf("metric not support")
	}
	err = c.handler.ModifyMetricMeta(meta)
	if err != nil {
		return nil, err
	}

	m, ok := c.MetricMap[meta.MetricName]
	if !ok {
		conf, err := metric.NewTcmMetricConfigWithMetricYaml(mconf, meta)
		if err != nil {
			return nil, err
		}
		nm, err := metric.NewTcmMetric(meta, conf)
		if err != nil {
			return nil, err
		}
		// 指标过滤
		if !c.handler.IsMetricValid(nm) {
			return nil, fmt.Errorf("metric not support")
		}
		err = c.handler.ModifyMetric(nm)
		if err != nil {
			return nil, err
		}
		return nm, nil
	}
	return m, nil
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
		numSeries += len(q.Metric.SeriesCache.Series)
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
			pms, err0 := q.GetPromMetrics()
			if err0 != nil {
				level.Error(c.logger).Log(
					"msg", "Get samples fail",
					"err", err,
					"metric", q.Metric.Id,
				)
				err = err0
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

type TcProductCollectorReloader struct {
	collector     *TcProductCollector
	relodInterval time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	logger        log.Logger
}

func (r *TcProductCollectorReloader) Run() {
	ticker := time.NewTicker(r.relodInterval)
	defer ticker.Stop()

	// sleep when first start
	time.Sleep(r.relodInterval)

	for {
		level.Info(r.logger).Log("msg", "start reload product metadata", "Namespace", r.collector.Namespace)
		e := r.reloadMetricsByProductConf()
		if e != nil {
			level.Error(r.logger).Log("msg", "reload product error", "err", e,
				"namespace", r.collector.Namespace)
		}
		level.Info(r.logger).Log("msg", "complete reload product metadata", "Namespace", r.collector.Namespace)
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *TcProductCollectorReloader) Stop() {
	r.cancel()
}

func (r *TcProductCollectorReloader) reloadMetricsByProductConf() error {
	return r.collector.LoadMetricsByProductConf()
}

// 创建新的TcProductCollector, 每个产品一个
func NewTcProductCollector(namespace string, metricRepo metric.TcmMetricRepository, cred common.CredentialIface,
	conf *config.TencentConfig, pconf *config.TencentProduct, logger log.Logger) (*TcProductCollector, error) {
	factory, exists := handlerFactoryMap[namespace]
	if !exists {
		return nil, fmt.Errorf("product handler not found, Namespace=%s ", namespace)
	}

	var instanceRepoCache instance.TcInstanceRepository
	if !util.IsStrInList(constant.NotSupportInstanceNamespaces, namespace) {
		// 支持实例自动发现的产品
		instanceRepo, err := instance.NewTcInstanceRepository(namespace, cred, conf, logger)
		if err != nil {
			return nil, err
		}
		// var instanceRepo instance.TcInstanceRepository
		// 使用instance缓存
		reloadInterval := time.Duration(pconf.RelodIntervalMinutes * int64(time.Minute))
		instanceRepoCache = instance.NewTcInstanceCache(instanceRepo, reloadInterval, logger)
	}

	c := &TcProductCollector{
		Namespace:    namespace,
		MetricRepo:   metricRepo,
		InstanceRepo: instanceRepoCache,
		Conf:         conf,
		ProductConf:  pconf,
		logger:       logger,
	}

	handler, err := factory(cred, c, logger)
	if err != nil {
		return nil, err
	}
	c.handler = handler

	err = c.LoadMetricsByMetricConf()
	if err != nil {
		return nil, err
	}
	start := time.Now()
	err = c.LoadMetricsByProductConf()
	if err != nil {
		return nil, err
	}
	fmt.Println("耗时", time.Since(start).Milliseconds())
	err = c.initQuerys()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewTcProductCollectorReloader(ctx context.Context, collector *TcProductCollector,
	relodInterval time.Duration, logger log.Logger) *TcProductCollectorReloader {
	childCtx, cancel := context.WithCancel(ctx)
	reloader := &TcProductCollectorReloader{
		collector:     collector,
		relodInterval: relodInterval,
		ctx:           childCtx,
		cancel:        cancel,
		logger:        logger,
	}
	return reloader
}
