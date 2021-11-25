package instance

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/memcached/v20190318"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/MEMCACHED", NewMemcachedTcInstanceRepository)
}

type MemcachedTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *MemcachedTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *MemcachedTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	req.InstanceIds = []*string{&id}
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.InstanceList) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.InstanceList[0]
	instance, err = NewMemcachedTcInstance(fmt.Sprintf("%d", *meta.CmemId), meta)
	if err != nil {
		return
	}
	return
}

func (repo *MemcachedTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *MemcachedTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total uint64 = 0

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if total == 0 {
		total = uint64(*resp.Response.TotalNum)
	}
	for _, meta := range resp.Response.InstanceList {
		ins, e := NewMemcachedTcInstance(fmt.Sprintf("%d", *meta.CmemId), meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create memcached instance fail", "id", *meta.InstanceId)
			continue
		}
		instances = append(instances, ins)
	}
	offset += limit
	if offset < total {
		req.Offset = &offset
		goto getMoreInstances
	}
	return
}

func NewMemcachedTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewMemcacheClient(c)
	if err != nil {
		return
	}
	repo = &MemcachedTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
