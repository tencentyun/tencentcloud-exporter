package instance

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	pkgcommon "github.com/tencentyun/tencentcloud-exporter/pkg/common"
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
	req.QueryType = common.StringPtr("zookeeper")
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

func (repo *ZookeeperTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *ZookeeperTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeSREInstancesRequest()
	req.QueryType = common.StringPtr("zookeeper")
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

// ZookeeperPod
type ZookeeperTcInstancePodRepository interface {
	GetZookeeperPodInfo(instanceId string) (*sdk.DescribeZookeeperReplicasResponse, error)
}

type ZookeeperTcInstancePodRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *ZookeeperTcInstancePodRepositoryImpl) GetZookeeperPodInfo(instanceId string) (*sdk.DescribeZookeeperReplicasResponse, error) {
	req := sdk.NewDescribeZookeeperReplicasRequest()
	req.InstanceId = common.StringPtr(instanceId)
	return repo.client.DescribeZookeeperReplicas(req)
}
func NewZookeeperTcInstancePodRepository(c *config.TencentConfig, logger log.Logger) (ZookeeperTcInstancePodRepository, error) {
	cli, err := client.NewTseClient(c)
	if err != nil {
		return nil, err
	}
	repo := &ZookeeperTcInstancePodRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

// ZookeeperInterface
type ZookeeperTcInstanceInterfaceRepository interface {
	GetZookeeperInterfaceInfo(instanceId string) (*sdk.DescribeZookeeperServerInterfacesResponse, error)
}

type ZookeeperTcInstanceInterfaceRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *ZookeeperTcInstanceInterfaceRepositoryImpl) GetZookeeperInterfaceInfo(instanceId string) (*sdk.DescribeZookeeperServerInterfacesResponse, error) {
	req := sdk.NewDescribeZookeeperServerInterfacesRequest()
	req.InstanceId = common.StringPtr(instanceId)
	return repo.client.DescribeZookeeperServerInterfaces(req)
}
func NewZookeeperTcInstanceInterfaceRepository(cred pkgcommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (ZookeeperTcInstanceInterfaceRepository, error) {
	cli, err := client.NewTseClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &ZookeeperTcInstanceInterfaceRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

func NewZookeeperTcInstanceRepository(cred pkgcommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewTseClient(cred, c)
	if err != nil {
		return
	}
	repo = &ZookeeperTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
