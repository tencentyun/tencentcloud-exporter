package instance

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/LIGHTHOUSE", NewLighthouseTcInstanceRepository)
}

type LighthouseTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *LighthouseTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *LighthouseTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	req.InstanceIds = []*string{&id}
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.InstanceSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.InstanceSet[0]
	instance, err = NewLighthouseTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *LighthouseTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *LighthouseTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
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
		ins, e := NewLighthouseTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create lighthouse instance fail", "id", *meta.InstanceId)
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

func NewLighthouseTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewLighthouseClient(c)
	if err != nil {
		return
	}
	repo = &LighthouseTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
