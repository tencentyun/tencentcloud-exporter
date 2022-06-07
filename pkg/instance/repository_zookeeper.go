package instance

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("TSE/ZOOKEEPER", NewZookeeperTcInstanceRepository)
}

type ZookeeperTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *ZookeeperTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *ZookeeperTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeSREInstancesRequest()
	// req.Filters.Name = []*string{&id}
	resp, err := repo.client.DescribeSREInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.Content) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.Content[0]
	instance, err = NewZookeeperTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *ZookeeperTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *ZookeeperTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeSREInstancesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeSREInstances(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.Content {
		ins, e := NewZookeeperTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create tdmq instance fail", "id", *meta.InstanceId)
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

func NewZookeeperTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewTseClient(c)
	if err != nil {
		return
	}
	repo = &ZookeeperTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
