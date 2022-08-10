package cachedtransactiongather

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/promlog"
	"sort"
	"sync"
	"testing"
	"time"
)

type mockGatherer struct {
	sleepUntil time.Duration
}

func (m mockGatherer) Gather() ([]*io_prometheus_client.MetricFamily, error) {
	fmt.Println("start gather: " + m.sleepUntil.String())
	time.Sleep(m.sleepUntil)
	fmt.Println("end gather: " + m.sleepUntil.String())
	return []*io_prometheus_client.MetricFamily{}, nil
}

func newMockGatherer(duration time.Duration) prometheus.Gatherer {
	return &mockGatherer{
		sleepUntil: duration,
	}
}

type multiTRegistry struct {
	tGatherers []prometheus.TransactionalGatherer
}

func newMultiConcurrencyRegistry(tGatherers ...prometheus.TransactionalGatherer) *multiTRegistry {
	return &multiTRegistry{
		tGatherers: tGatherers,
	}
}

// Gather implements TransactionalGatherer interface.
func (r *multiTRegistry) Gather() (mfs []*io_prometheus_client.MetricFamily, done func(), err error) {
	dFns := make([]func(), 0, len(r.tGatherers))
	wait := sync.WaitGroup{}
	wait.Add(len(r.tGatherers))
	for i := range r.tGatherers {
		go func(i int) {
			_, _, _ = r.tGatherers[i].Gather()
			wait.Done()
		}(i)
	}
	wait.Wait()

	sort.Slice(mfs, func(i, j int) bool {
		return *mfs[i].Name < *mfs[j].Name
	})
	return mfs, func() {
		for _, d := range dFns {
			d()
		}
	}, nil
}

func TestCache(t *testing.T) {
	promlogConfig := &promlog.Config{}
	cacheInterval := 60 * time.Second
	logger := promlog.New(promlogConfig)
	gather := NewCachedTransactionGather(
		newMultiConcurrencyRegistry(
			prometheus.ToTransactionalGatherer(newMockGatherer(time.Second*40)),
			prometheus.ToTransactionalGatherer(newMockGatherer(time.Second*23)),
			prometheus.ToTransactionalGatherer(newMockGatherer(time.Second*7)),
		),
		cacheInterval, logger,
	)

	t.Run("gather with multiple calls should not error", func(t *testing.T) {
		wait := sync.WaitGroup{}
		wait.Add(10)
		for range [10]int{} {
			go func() {
				begin := time.Now()
				mfs, done, err := gather.Gather()
				defer done()
				if err != nil {
					logger.Log("err", err)
					t.Errorf("gather error: %v", err)
				}
				logger.Log("mfs", mfs, "done", "err", err)
				if time.Since(begin) > cacheInterval {
					t.Errorf("gather cost more than cacheInterval %v", time.Since(begin).String())
				}
				wait.Done()
			}()
		}
		wait.Wait()
	})

	t.Run("gather success", func(t *testing.T) {
		wait := sync.WaitGroup{}
		wait.Add(3)
		go func() {
			mfs, done, err := gather.Gather()
			defer done()
			if err != nil {
				logger.Log("err", err)
				t.Errorf("gather error: %v", err)
			}
			logger.Log("mfs", mfs, "done", "err", err)
			wait.Done()
		}()
		go func() {
			mfs, done, err := gather.Gather()
			defer done()
			if err != nil {
				logger.Log("err", err)
				t.Errorf("gather error: %v", err)
			}
			logger.Log("mfs", mfs, "done", "err", err)
			wait.Done()
		}()
		go func() {
			mfs, done, err := gather.Gather()
			defer done()
			if err != nil {
				logger.Log("err", err)
				t.Errorf("gather error: %v", err)
			}
			logger.Log("mfs", mfs, "done", "err", err)
			wait.Done()
		}()
		wait.Wait()
	})
}
