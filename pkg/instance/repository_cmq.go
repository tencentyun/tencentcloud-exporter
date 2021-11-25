package instance

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cmq/v20190304"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/CMQ", NewCMQTcInstanceRepository)
}

type CMQTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *CMQTcInstanceRepository) GetInstanceKey() string {
	return "QueueId"
}

func (repo *CMQTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeQueueDetailRequest()
	req.QueueName = &id
	resp, err := repo.client.DescribeQueueDetail(req)
	if err != nil {
		return
	}
	if len(resp.Response.QueueSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.QueueSet[0]
	instance, err = NewCMQTcInstance(*meta.QueueId, meta)
	if err != nil {
		return
	}
	return
}

func (repo *CMQTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *CMQTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeQueueDetailRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total uint64 = 0

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeQueueDetail(req)
	if err != nil {
		return
	}
	if total == 0 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.QueueSet {
		ins, e := NewCMQTcInstance(*meta.QueueId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create cmq instance fail", "id", *meta.QueueId)
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

func NewCMQTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewCMQClient(c)
	if err != nil {
		return
	}
	repo = &CMQTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
