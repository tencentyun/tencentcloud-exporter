package instance

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mongodb/v20190725"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
	"strconv"
)

func init() {
	registerRepository("QCE/CMONGO", NewMongoTcInstanceRepository)
}

type MongoTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *MongoTcInstanceRepository) GetInstanceKey() string {
	return "target"
}

func (repo *MongoTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	req.InstanceIds = []*string{&id}
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.InstanceDetails) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.InstanceDetails[0]
	instance, err = NewMongoTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *MongoTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *MongoTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Limit = &limit

	if v, ok := filters["ProjectId"]; ok {
		tv, e := strconv.ParseInt(v, 10, 64)
		utv := uint64(tv)
		if e == nil {
			req.ProjectIds = []*uint64{&utv}
		}
	}
	if v, ok := filters["InstanceId"]; ok {
		req.InstanceIds = []*string{&v}
	}

getMoreInstances:
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.InstanceDetails {
		ins, e := NewMongoTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create mongo instance fail", "id", *meta.InstanceId)
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

func NewMongoTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewMongodbClient(c)
	if err != nil {
		return
	}
	repo = &MongoTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
