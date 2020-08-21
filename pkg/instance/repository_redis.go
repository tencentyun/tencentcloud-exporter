package instance

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/REDIS", NewRedisTcInstanceRepository)
}

type RedisTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *RedisTcInstanceRepository) GetInstanceKey() string {
	return "instanceid"
}

func (repo *RedisTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	req.InstanceId = &id
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.InstanceSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.InstanceSet[0]
	instance, err = NewRedisTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *RedisTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *RedisTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.InstanceSet {
		ins, e := NewRedisTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create redis instance fail", "id", *meta.InstanceId)
			continue
		}
		instances = append(instances, ins)
	}
	offset += limit
	if offset < uint64(total) {
		goto getMoreInstances
	}

	return
}

func NewRedisTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewRedisClient(c)
	if err != nil {
		return
	}
	repo = &RedisTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
