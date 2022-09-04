package cachedtransactiongather

import (
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

// NewCachedTransactionGather creates cachedTransactionGather, which only allow one request per interval, and pending others
// until response
func NewCachedTransactionGather(
	gather prometheus.TransactionalGatherer,
	cacheInterval time.Duration,
	logger log.Logger,
) prometheus.TransactionalGatherer {
	return &cachedTransactionGather{
		gather:             gather,
		nextCollectionTime: time.Now(),
		cacheInterval:      cacheInterval,
		logger:             logger,
	}
}

type cachedTransactionGather struct {
	gather prometheus.TransactionalGatherer

	cache []*io_prometheus_client.MetricFamily
	err   error

	nextCollectionTime time.Time
	cacheInterval      time.Duration

	lock sync.RWMutex

	logger log.Logger
}

func (c *cachedTransactionGather) Gather() ([]*io_prometheus_client.MetricFamily, func(), error) {
	c.lock.Lock()
	shouldGather := time.Now().After(c.nextCollectionTime)
	if shouldGather {
		begin := time.Now()
		c.nextCollectionTime = c.nextCollectionTime.Add(c.cacheInterval)
		metrics, done, err := c.gather.Gather()
		if err != nil {
			c.err = err
			c.cache = []*io_prometheus_client.MetricFamily{}
			done()
		} else {
			c.cache = metrics
			c.err = nil
			done()
		}
		c.lock.Unlock()
		duration := time.Since(begin)
		level.Info(c.logger).Log("msg", "Collect all products done", "duration_seconds", duration.Seconds())
	} else {
		c.lock.Unlock()
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.err != nil {
		return nil, func() {}, c.err
	}
	return c.cache, func() {}, nil
}
