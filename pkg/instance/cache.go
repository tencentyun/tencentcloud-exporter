package instance

import (
	"fmt"
	dts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dts/v20180330"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	dtsNew "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dts/v20211206"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	tdmq "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"
	tse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"
	vbc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
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

// tdmq
type TcTdmqInstanceNamespaceCache struct {
	Raw            TdmqTcInstanceRocketMQNameSpacesRepository
	cache          map[string]*tdmq.DescribeRocketMQNamespacesResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcTdmqInstanceNamespaceCache) GetRocketMQNamespacesInfo(instanceId string) (*tdmq.DescribeRocketMQNamespacesResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		namespace, ok := c.cache[instanceId]
		if ok {
			return namespace, nil
		}
	}

	namespace, err := c.Raw.GetRocketMQNamespacesInfo(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = namespace
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get RocketMQ Namespaces info from api", "instanceId", instanceId)
	return namespace, nil
}

func NewTcTdmqInstanceNamespaceCache(repo TdmqTcInstanceRocketMQNameSpacesRepository, reloadInterval time.Duration, logger log.Logger) TdmqTcInstanceRocketMQNameSpacesRepository {
	cache := &TcTdmqInstanceNamespaceCache{
		Raw:            repo,
		cache:          map[string]*tdmq.DescribeRocketMQNamespacesResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

type TcTdmqInstanceTopicsCache struct {
	Raw            TdmqTcInstanceRocketMQTopicsRepository
	cache          map[string]*tdmq.DescribeRocketMQTopicsResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcTdmqInstanceTopicsCache) GetRocketMQTopicsInfo(instanceId string, namespaceId string) (*tdmq.DescribeRocketMQTopicsResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		topic, ok := c.cache[instanceId]
		if ok {
			return topic, nil
		}
	}

	topic, err := c.Raw.GetRocketMQTopicsInfo(instanceId, namespaceId)
	if err != nil {
		return nil, err
	}
	instanceIdNamspace := fmt.Sprintf("%v-%v", instanceId, namespaceId)
	c.cache[instanceIdNamspace] = topic
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get RocketMQ Namespaces info from api", "instanceId", instanceId)
	return topic, nil
}

func NewTcTdmqInstanceTopicsCache(repo TdmqTcInstanceRocketMQTopicsRepository, reloadInterval time.Duration, logger log.Logger) TdmqTcInstanceRocketMQTopicsRepository {
	cache := &TcTdmqInstanceTopicsCache{
		Raw:            repo,
		cache:          map[string]*tdmq.DescribeRocketMQTopicsResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

// zookeeper
type TcZookeeperInstancePodCache struct {
	Raw            ZookeeperTcInstancePodRepository
	cache          map[string]*tse.DescribeZookeeperReplicasResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcZookeeperInstancePodCache) GetZookeeperPodInfo(instanceId string) (*tse.DescribeZookeeperReplicasResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		namespace, ok := c.cache[instanceId]
		if ok {
			return namespace, nil
		}
	}

	pod, err := c.Raw.GetZookeeperPodInfo(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = pod
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get RocketMQ Namespaces info from api", "instanceId", instanceId)
	return pod, nil
}

func NewTcZookeeperInstancePodCache(repo ZookeeperTcInstancePodRepository, reloadInterval time.Duration, logger log.Logger) ZookeeperTcInstancePodRepository {
	cache := &TcZookeeperInstancePodCache{
		Raw:            repo,
		cache:          map[string]*tse.DescribeZookeeperReplicasResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

type TcZookeeperInstanceInterfaceCache struct {
	Raw            ZookeeperTcInstanceInterfaceRepository
	cache          map[string]*tse.DescribeZookeeperServerInterfacesResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcZookeeperInstanceInterfaceCache) GetZookeeperInterfaceInfo(instanceId string) (*tse.DescribeZookeeperServerInterfacesResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		topic, ok := c.cache[instanceId]
		if ok {
			return topic, nil
		}
	}

	interfaceInfo, err := c.Raw.GetZookeeperInterfaceInfo(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = interfaceInfo
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get RocketMQ Namespaces info from api", "instanceId", instanceId)
	return interfaceInfo, nil
}

func NewTcZookeeperInstanceInterfaceCache(repo ZookeeperTcInstanceInterfaceRepository, reloadInterval time.Duration, logger log.Logger) ZookeeperTcInstanceInterfaceRepository {
	cache := &TcZookeeperInstanceInterfaceCache{
		Raw:            repo,
		cache:          map[string]*tse.DescribeZookeeperServerInterfacesResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

// nacos
type TcNacosInstancePodCache struct {
	Raw            NacosTcInstancePodRepository
	cache          map[string]*tse.DescribeNacosReplicasResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcNacosInstancePodCache) GetNacosPodInfo(instanceId string) (*tse.DescribeNacosReplicasResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		namespace, ok := c.cache[instanceId]
		if ok {
			return namespace, nil
		}
	}

	pod, err := c.Raw.GetNacosPodInfo(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = pod
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get RocketMQ Namespaces info from api", "instanceId", instanceId)
	return pod, nil
}

func NewTcNacosInstancePodCache(repo NacosTcInstancePodRepository, reloadInterval time.Duration, logger log.Logger) NacosTcInstancePodRepository {
	cache := &TcNacosInstancePodCache{
		Raw:            repo,
		cache:          map[string]*tse.DescribeNacosReplicasResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

type TcNacosInstanceInterfaceCache struct {
	Raw            NacosTcInstanceInterfaceRepository
	cache          map[string]*tse.DescribeNacosServerInterfacesResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcNacosInstanceInterfaceCache) GetNacosInterfaceInfo(instanceId string) (*tse.DescribeNacosServerInterfacesResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		topic, ok := c.cache[instanceId]
		if ok {
			return topic, nil
		}
	}

	interfaceInfo, err := c.Raw.GetNacosInterfaceInfo(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = interfaceInfo
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get RocketMQ Namespaces info from api", "instanceId", instanceId)
	return interfaceInfo, nil
}

func NewTcNacosInstanceInterfaceCache(repo NacosTcInstanceInterfaceRepository, reloadInterval time.Duration, logger log.Logger) NacosTcInstanceInterfaceRepository {
	cache := &TcNacosInstanceInterfaceCache{
		Raw:            repo,
		cache:          map[string]*tse.DescribeNacosServerInterfacesResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

// dts
type TcDtsInstanceMigrateInfosCache struct {
	Raw            DtsTcInstanceMigrateInfosRepository
	cache          map[string]*dts.DescribeMigrateJobsResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcDtsInstanceMigrateInfosCache) GetMigrateInfos(instanceId string) (*dts.DescribeMigrateJobsResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		namespace, ok := c.cache[instanceId]
		if ok {
			return namespace, nil
		}
	}

	migrateInfos, err := c.Raw.GetMigrateInfos(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = migrateInfos
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get dts Namespaces info from api", "instanceId", instanceId)
	return migrateInfos, nil
}

func NewTcDtsInstanceMigrateInfosCache(repo DtsTcInstanceMigrateInfosRepository, reloadInterval time.Duration, logger log.Logger) DtsTcInstanceMigrateInfosRepository {
	cache := &TcDtsInstanceMigrateInfosCache{
		Raw:            repo,
		cache:          map[string]*dts.DescribeMigrateJobsResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

type TcDtsInstanceReplicationCache struct {
	Raw            DtsTcInstanceReplicationsRepository
	cache          map[string]*dtsNew.DescribeSyncJobsResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcDtsInstanceReplicationCache) GetReplicationsInfo(instanceId string) (*dtsNew.DescribeSyncJobsResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		namespace, ok := c.cache[instanceId]
		if ok {
			return namespace, nil
		}
	}

	replicationsInfo, err := c.Raw.GetReplicationsInfo(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = replicationsInfo
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get dts Namespaces info from api", "instanceId", instanceId)
	return replicationsInfo, nil
}

func NewTcDtsInstanceReplicationsInfosCache(repo DtsTcInstanceReplicationsRepository, reloadInterval time.Duration, logger log.Logger) DtsTcInstanceReplicationsRepository {
	cache := &TcDtsInstanceReplicationCache{
		Raw:            repo,
		cache:          map[string]*dtsNew.DescribeSyncJobsResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

// vbc
type TcVbcInstanceeDRegionCache struct {
	Raw            VbcTcInstanceDRegionRepository
	cache          map[string]*vbc.DescribeCcnRegionBandwidthLimitsResponse
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcVbcInstanceeDRegionCache) GetVbcDRegionInfo(instanceId string) (*vbc.DescribeCcnRegionBandwidthLimitsResponse, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		namespace, ok := c.cache[instanceId]
		if ok {
			return namespace, nil
		}
	}

	dRegion, err := c.Raw.GetVbcDRegionInfo(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = dRegion
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get vbc Namespaces info from api", "instanceId", instanceId)
	return dRegion, nil
}

func NewVbcTcInstanceDRegionRepositoryCache(repo VbcTcInstanceDRegionRepository, reloadInterval time.Duration, logger log.Logger) VbcTcInstanceDRegionRepository {
	cache := &TcVbcInstanceeDRegionCache{
		Raw:            repo,
		cache:          map[string]*vbc.DescribeCcnRegionBandwidthLimitsResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}
