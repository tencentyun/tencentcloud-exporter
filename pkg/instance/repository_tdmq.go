package instance

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
	req := sdk.NewDescribeAMQPClustersRequest()
	req.ClusterIdList = []*string{&id}
	resp, err := repo.client.DescribeAMQPClusters(req)
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
	req := sdk.NewDescribeAMQPClustersRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeAMQPClusters(req)
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

func NewTdmqTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewTDMQClient(c)
	if err != nil {
		return
	}
	repo = &TdmqTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
