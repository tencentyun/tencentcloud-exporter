package instance

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/postgres/v20170312"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

var  idKey = "db-instance-id"

func init() {
	registerRepository("QCE/POSTGRES", NewPGTcInstanceRepository)
}

type PGTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *PGTcInstanceRepository) GetInstanceKey() string {
	return "DBInstanceId"
}

func (repo *PGTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	req.Filters =[]*sdk.Filter{{
		Name:  &idKey,
		Values: []*string{&id},
	}}
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.DBInstanceSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.DBInstanceSet[0]
	instance, err = NewPGTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *PGTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *PGTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total uint64 = 0

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if total == 0 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.DBInstanceSet{
		ins, e := NewPGTcInstance(*meta.DBInstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create pg instance fail", "id", *meta.DBInstanceId)
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

func NewPGTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewPGClient(c)
	if err != nil {
		return
	}
	repo = &PGTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
