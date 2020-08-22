package instance

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"sync"
	"time"
)

type TcInstanceCache struct {
	Raw            TcInstanceRepository
	cache          map[string]TcInstance
	lastReloadTime int64
	logger         log.Logger
	mu             sync.Mutex
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

func (c *TcInstanceCache) checkNeedreload() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lastReloadTime != 0 {
		return nil
	}

	inss, err := c.Raw.ListByFilters(map[string]string{})
	if err != nil {
		return err
	}
	for _, instance := range inss {
		c.cache[instance.GetInstanceId()] = instance
	}
	c.lastReloadTime = time.Now().Unix()
	level.Info(c.logger).Log("msg", "Reload instance cache", "num", len(c.cache))
	return
}

func NewTcInstanceCache(repo TcInstanceRepository, logger log.Logger) TcInstanceRepository {
	cache := &TcInstanceCache{
		Raw:    repo,
		cache:  map[string]TcInstance{},
		logger: logger,
	}
	return cache

}
