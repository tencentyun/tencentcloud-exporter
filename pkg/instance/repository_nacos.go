package instance

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("TSE/NACOS", NewNaocsTcInstanceRepository)
}

type NacosTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *NacosTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *NacosTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeSREInstancesRequest()
	req.QueryType = common.StringPtr("nacos")
	req.QuerySource = common.StringPtr("cloud_metrics")
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

func (repo *NacosTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *NacosTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeSREInstancesRequest()
	req.QueryType = common.StringPtr("nacos")
	req.QuerySource = common.StringPtr("cloud_metrics")
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

// NacosPod
type NacosTcInstancePodRepository interface {
	GetNacosPodInfo(instanceId string) (*sdk.DescribeNacosReplicasResponse, error)
}

type NacosTcInstancePodRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *NacosTcInstancePodRepositoryImpl) GetNacosPodInfo(instanceId string) (*sdk.DescribeNacosReplicasResponse, error) {
	req := sdk.NewDescribeNacosReplicasRequest()
	req.InstanceId = common.StringPtr(instanceId)
	return repo.client.DescribeNacosReplicas(req)
}
func NewNacosTcInstancePodRepository(c *config.TencentConfig, logger log.Logger) (NacosTcInstancePodRepository, error) {
	cli, err := client.NewTseClient(c)
	if err != nil {
		return nil, err
	}
	repo := &NacosTcInstancePodRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

// NacosInterface
type NacosTcInstanceInterfaceRepository interface {
	GetNacosInterfaceInfo(instanceId string) (*sdk.DescribeNacosServerInterfacesResponse, error)
}

type NacosTcInstanceInterfaceRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *NacosTcInstanceInterfaceRepositoryImpl) GetNacosInterfaceInfo(instanceId string) (*sdk.DescribeNacosServerInterfacesResponse, error) {
	req := sdk.NewDescribeNacosServerInterfacesRequest()
	req.InstanceId = common.StringPtr(instanceId)
	return repo.client.DescribeNacosServerInterfaces(req)
}
func NewNacosTcInstanceInterfaceRepository(c *config.TencentConfig, logger log.Logger) (NacosTcInstanceInterfaceRepository, error) {
	cli, err := client.NewTseClient(c)
	if err != nil {
		return nil, err
	}
	repo := &NacosTcInstanceInterfaceRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

func NewNaocsTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewTseClient(c)
	if err != nil {
		return
	}
	repo = &NacosTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
