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
	registerRepository("QCE/CMQTOPIC", NewCMQTopicTcInstanceRepository)
}

type CMQTopicTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *CMQTopicTcInstanceRepository) GetInstanceKey() string {
	return "TopicId"
}

func (repo *CMQTopicTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeTopicDetailRequest()
	req.TopicName = &id
	resp, err := repo.client.DescribeTopicDetail(req)
	if err != nil {
		return
	}
	if len(resp.Response.TopicSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.TopicSet[0]
	instance, err = NewCMQTopicTcInstance(*meta.TopicId, meta)
	if err != nil {
		return
	}
	return
}

func (repo *CMQTopicTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *CMQTopicTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeTopicDetailRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total uint64 = 0

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeTopicDetail(req)
	if err != nil {
		return
	}
	if total == 0 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.TopicSet {
		ins, e := NewCMQTopicTcInstance(*meta.TopicId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create cmq topic fail", "id", *meta.TopicId)
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

func NewCMQTopicTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewCMQClient(c)
	if err != nil {
		return
	}
	repo = &CMQTopicTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
