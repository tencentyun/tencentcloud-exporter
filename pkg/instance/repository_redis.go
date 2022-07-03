package instance

import (
	"fmt"

	mycommon "github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

//go:generate mockgen -source=./repository_redis.go -destination=./repository_redis_mock.go -package=instance

func init() {
	registerRepository("QCE/REDIS", NewRedisTcInstanceRepository)
	registerRepository("QCE/REDIS_MEM", NewRedisTcInstanceRepository)
}

type RedisTcInstanceRepository struct {
	credential mycommon.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func NewRedisTcInstanceRepository(cred mycommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewRedisClient(cred, c)
	if err != nil {
		return
	}
	repo = &RedisTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
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
	req.Offset = &offset
	req.Limit = &limit
	req.Status = []*int64{common.Int64Ptr(2)}

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
		req.Offset = &offset
		goto getMoreInstances
	}

	return
}

type RedisTcInstanceNodeRepository interface {
	GetNodeInfo(instanceId string) (*sdk.DescribeInstanceNodeInfoResponse, error)
}

type RedisTcInstanceNodeRepositoryImpl struct {
	credential mycommon.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *RedisTcInstanceNodeRepositoryImpl) GetNodeInfo(instanceId string) (*sdk.DescribeInstanceNodeInfoResponse, error) {
	req := sdk.NewDescribeInstanceNodeInfoRequest()
	req.InstanceId = common.StringPtr(instanceId)
	return repo.client.DescribeInstanceNodeInfo(req)
}

func NewRedisTcInstanceNodeRepository(cred mycommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (RedisTcInstanceNodeRepository, error) {
	cli, err := client.NewRedisClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &RedisTcInstanceNodeRepositoryImpl{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return repo, nil
}
