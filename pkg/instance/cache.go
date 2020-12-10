package instance

import (
	"sync"
	"time"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"

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
	ins, exists := c.cache[id]
	if exists {
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

type TcRedisInstanceNodeCache struct {
	Raw            RedisTcInstanceNodeRepository
	cache          map[string]*sdk.DescribeInstanceNodeInfoResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcRedisInstanceNodeCache) GetNodeInfo(instanceId string) (*sdk.DescribeInstanceNodeInfoResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		node, ok := c.cache[instanceId]
		if ok {
			return node, nil
		}
	}

	node, err := c.Raw.GetNodeInfo(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = node
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get redis node info from api", "instanceId", instanceId)
	return node, nil
}

func NewTcRedisInstanceNodeCache(repo RedisTcInstanceNodeRepository, reloadInterval time.Duration, logger log.Logger) RedisTcInstanceNodeRepository {
	cache := &TcRedisInstanceNodeCache{
		Raw:            repo,
		cache:          map[string]*sdk.DescribeInstanceNodeInfoResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}
