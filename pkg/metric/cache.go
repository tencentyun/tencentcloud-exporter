package metric

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// 腾讯云监控指标缓存, 在TcmMetricRepository封装一层, 指标元数据使用缓存, 转发获取数据点请求
type TcmMetricCache struct {
	Raw                TcmMetricRepository
	metaCache          map[string]map[string]*TcmMeta //k1=namespace, k2=metricname(小写)
	metaLastReloadTime map[string]int64
	logger             log.Logger
}

func (c *TcmMetricCache) GetMeta(namespace string, name string) (*TcmMeta, error) {
	err := c.checkMetaNeedreload(namespace)
	if err != nil {
		return nil, err
	}
	np, exists := c.metaCache[namespace]
	if !exists {
		return nil, fmt.Errorf("namespace cache not exists")
	}
	m, exists := np[strings.ToLower(name)]
	if !exists {
		return nil, fmt.Errorf("metric cache not exists")
	}
	return m, nil
}

func (c *TcmMetricCache) ListMetaByNamespace(namespace string) ([]*TcmMeta, error) {
	err := c.checkMetaNeedreload(namespace)
	if err != nil {
		return nil, err
	}
	var metas []*TcmMeta
	for _, meta := range c.metaCache[namespace] {
		metas = append(metas, meta)
	}
	return metas, nil
}

func (c *TcmMetricCache) GetSamples(series *TcmSeries, startTime int64, endTime int64) (samples *TcmSamples, err error) {
	return c.Raw.GetSamples(series, startTime, endTime)
}

func (c *TcmMetricCache) ListSamples(metric *TcmMetric, startTime int64, endTime int64) (samplesList []*TcmSamples, err error) {
	return c.Raw.ListSamples(metric, startTime, endTime)
}

// 检测是否需要reload缓存的数据
func (c *TcmMetricCache) checkMetaNeedreload(namespace string) (err error) {
	v, ok := c.metaLastReloadTime[namespace]
	if ok && v != 0 {
		return nil
	}
	metas, err := c.Raw.ListMetaByNamespace(namespace)
	if err != nil {
		return err
	}
	np, ok := c.metaCache[namespace]
	if !ok {
		np = map[string]*TcmMeta{}
		c.metaCache[namespace] = np
	}
	for _, meta := range metas {
		np[strings.ToLower(meta.MetricName)] = meta
	}
	c.metaLastReloadTime[namespace] = time.Now().Unix()

	level.Info(c.logger).Log("msg", "Reload metric meta cache", "namespace", namespace, "num", len(np))
	return
}

func NewTcmMetricCache(repo TcmMetricRepository, logger log.Logger) TcmMetricRepository {
	cache := &TcmMetricCache{
		Raw:                repo,
		metaCache:          map[string]map[string]*TcmMeta{},
		metaLastReloadTime: map[string]int64{},
		logger:             logger,
	}
	return cache

}
