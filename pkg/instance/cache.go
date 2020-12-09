package instance

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	DefaultReloadInterval = 60 * time.Minute
)

// 可用于产品的实例的缓存, TcInstanceRepository
type TcInstanceCache struct {
	Raw            TcInstanceRepository
	cache          map[string]TcInstance
	lastReloadTime time.Time
	logger         log.Logger
	mu             sync.Mutex
	reloadInterval time.Duration
}

func (c *TcInstanceCache) GetInstanceKey() string {
	return c.Raw.GetInstanceKey()
}

func (c *TcInstanceCache) Get(id string) (TcInstance, error) {
	ins, ok := c.cache[id]
	if ok {
		return ins, nil
	}

	ins, err := c.Raw.Get(id)
	if err != nil {
		return nil, err
	}
	c.cache[ins.GetInstanceId()] = ins
	return ins, nil
}

func (c *TcInstanceCache) ListByIds(ids []string) (insList []TcInstance, err error) {
	err = c.checkNeedreload()
	if err != nil {
		return nil, err
	}

	var notexists []string
	for _, id := range ids {
		ins, ok := c.cache[id]
		if ok {
			insList = append(insList, ins)
		} else {
			notexists = append(notexists, id)
		}
	}
	return
}

func (c *TcInstanceCache) ListByFilters(filters map[string]string) (insList []TcInstance, err error) {
	err = c.checkNeedreload()
	if err != nil {
		return
	}

	for _, ins := range c.cache {
		for k, v := range filters {
			tv, e := ins.GetFieldValueByName(k)
			if e != nil {
				break
			}
			if v != tv {
				break
			}
		}
		insList = append(insList, ins)
	}

	return
}

func (c *TcInstanceCache) checkNeedreload() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.lastReloadTime.IsZero() && time.Now().Sub(c.lastReloadTime) < c.reloadInterval {
		return nil
	}

	inss, err := c.Raw.ListByFilters(map[string]string{})
	if err != nil {
		return err
	}
	numChanged := 0
	if len(inss) > 0 {
		newCache := map[string]TcInstance{}
		for _, instance := range inss {
			newCache[instance.GetInstanceId()] = instance
		}
		numChanged = len(newCache) - len(c.cache)
		c.cache = newCache
	}
	c.lastReloadTime = time.Now()

	level.Info(c.logger).Log("msg", "Reload instance cache", "num", len(c.cache), "changed", numChanged)
	return nil
}

func NewTcInstanceCache(repo TcInstanceRepository, reloadInterval time.Duration, logger log.Logger) TcInstanceRepository {
	cache := &TcInstanceCache{
		Raw:            repo,
		cache:          map[string]TcInstance{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache

}
