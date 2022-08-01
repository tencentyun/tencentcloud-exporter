package instance

import (
	"fmt"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ckafka/v20190819"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/CKAFKA", NewKafkaTcInstanceRepository)
}

type KafkaTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *KafkaTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *KafkaTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	req.InstanceId = &id
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.Result.InstanceList) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.Result.InstanceList[0]
	instance, err = NewKafkaTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *KafkaTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *KafkaTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = 0

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if total == 0 {
		total = *resp.Response.Result.TotalCount
	}
	for _, meta := range resp.Response.Result.InstanceList {
		ins, e := NewKafkaTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create kafka instance fail", "id", *meta.InstanceId)
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

func NewKafkaTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewKafkaClient(cred, c)
	if err != nil {
		return
	}
	repo = &KafkaTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
