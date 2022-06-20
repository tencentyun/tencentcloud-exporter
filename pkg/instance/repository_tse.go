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
	registerRepository("TSE/ZOOKEEPER", NewTseTcInstanceRepository)
	registerRepository("TSE/NACOS", NewTseTcInstanceRepository)
}

type TseTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *TseTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *TseTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
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
	instance, err = NewTseTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *TseTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *TseTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
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
		ins, e := NewTseTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create tse instance fail", "id", *meta.InstanceId)
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

// type TseTcInstanceZookeeperRepository interface {
// 	GetZookeeperPodInfo(instanceId string) (*sdk.DescribeZookeeperReplicasResponse, error)
// }
//
// type RedisTcInstanceZookeeperRepositoryImpl struct {
// 	client *sdk.Client
// 	logger log.Logger
// }
//
// func (repo *RedisTcInstanceZookeeperRepositoryImpl) GetZookeeperPodInfo(instanceId string) (*sdk.DescribeZookeeperReplicasResponse, error) {
// 	req := sdk.NewDescribeZookeeperReplicasRequest()
// 	req.InstanceId = common.StringPtr(instanceId)
// 	return repo.client.DescribeZookeeperReplicas(req)
// }
//
// type TseTcInstanceNacosRepository interface {
// 	GetZookeeperPodInfo(instanceId string) (*sdk.DescribeNacosReplicasResponse, error)
// }
//
// type RedisTcInstanceNacosRepositoryImpl struct {
// 	client *sdk.Client
// 	logger log.Logger
// }
//
// func (repo *RedisTcInstanceNacosRepositoryImpl) GetNacosPodInfo(instanceId string) (*sdk.DescribeNacosReplicasResponse, error) {
// 	req := sdk.NewDescribeNacosReplicasRequest()
// 	req.InstanceId = common.StringPtr(instanceId)
// 	return repo.client.DescribeNacosReplicas(req)
// }

func NewTseTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewTseClient(c)
	if err != nil {
		return
	}
	repo = &TseTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
