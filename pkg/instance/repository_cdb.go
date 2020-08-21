package instance

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/CDB", NewCdbTcInstanceRepository)
}

type CdbTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *CdbTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *CdbTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	req.InstanceIds = []*string{&id}
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.Items) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.Items[0]
	instance, err = NewCdbTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *CdbTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *CdbTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	var offset uint64 = 0
	var limit uint64 = 2000
	var total int64 = -1

	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.Items {
		ins, e := NewCdbTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create cdb instance fail", "id", *meta.InstanceId)
			continue
		}
		instances = append(instances, ins)
	}
	offset += limit
	if offset < uint64(total) {
		goto getMoreInstances
	}

	return
}

func NewCdbTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewCdbClient(c)
	if err != nil {
		return
	}
	repo = &CdbTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
