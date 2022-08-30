package instance

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	selfcommon "github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/TDMQ", NewTdmqTcInstanceRepository)
}

type TdmqTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *TdmqTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *TdmqTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeRocketMQClustersRequest()
	req.ClusterIdList = []*string{&id}
	resp, err := repo.client.DescribeRocketMQClusters(req)
	if err != nil {
		return
	}
	if len(resp.Response.ClusterList) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.ClusterList[0]
	instance, err = NewTdmqTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *TdmqTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *TdmqTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeRocketMQClustersRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeRocketMQClusters(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.ClusterList {
		ins, e := NewTdmqTcInstance(*meta.Info.ClusterId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create tdmq instance fail", "id", *meta.Info.ClusterId)
			continue
		}
		instances = append(instances, ins)
	}
	offset += limit
	if offset < uint64(total) {
		req.Offset = &offset
		goto getMoreInstances
	}

	return
}

// RocketMQNameSpaces
type TdmqTcInstanceRocketMQNameSpacesRepository interface {
	GetRocketMQNamespacesInfo(instanceId string) (*sdk.DescribeRocketMQNamespacesResponse, error)
}

type TdmqTcInstanceRocketMQNameSpacesRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *TdmqTcInstanceRocketMQNameSpacesRepositoryImpl) GetRocketMQNamespacesInfo(instanceId string) (*sdk.DescribeRocketMQNamespacesResponse, error) {
	req := sdk.NewDescribeRocketMQNamespacesRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	req.Limit = &limit
	req.Offset = &offset
	req.ClusterId = common.StringPtr(instanceId)
	return repo.client.DescribeRocketMQNamespaces(req)
}

func NewTdmqTcInstanceRocketMQNameSpacesRepository(cred selfcommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (TdmqTcInstanceRocketMQNameSpacesRepository, error) {
	cli, err := client.NewTDMQClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &TdmqTcInstanceRocketMQNameSpacesRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

// RocketMQTopics
type TdmqTcInstanceRocketMQTopicsRepository interface {
	GetRocketMQTopicsInfo(instanceId string, namespaceId string) (*sdk.DescribeRocketMQTopicsResponse, error)
}

type TdmqTcInstanceRocketMQTopicsRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *TdmqTcInstanceRocketMQTopicsRepositoryImpl) GetRocketMQTopicsInfo(instanceId string, namespaceId string) (*sdk.DescribeRocketMQTopicsResponse, error) {
	req := sdk.NewDescribeRocketMQTopicsRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	req.Limit = &limit
	req.Offset = &offset
	req.ClusterId = common.StringPtr(instanceId)
	req.NamespaceId = common.StringPtr(namespaceId)
	return repo.client.DescribeRocketMQTopics(req)
}

func NewTdmqTcInstanceRocketMQTopicsRepository(cred selfcommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (TdmqTcInstanceRocketMQTopicsRepository, error) {
	cli, err := client.NewTDMQClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &TdmqTcInstanceRocketMQTopicsRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

func NewTdmqTcInstanceRepository(cred selfcommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewTDMQClient(cred, c)
	if err != nil {
		return
	}
	repo = &TdmqTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
