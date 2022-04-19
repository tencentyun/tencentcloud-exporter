package instance

import (
	"fmt"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sqlserver/v20180328"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/SQLSERVER", NewSqlServerTcInstanceRepository)
}

type SqlServerTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *SqlServerTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *SqlServerTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	req.InstanceIdSet = []*string{&id}
	repo.credential.Refresh()
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.DBInstances) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.DBInstances[0]
	instance, err = NewSqlServerTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *SqlServerTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *SqlServerTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	repo.credential.Refresh()
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.DBInstances {
		ins, e := NewSqlServerTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create cdb instance fail", "id", *meta.InstanceId)
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

func NewSqlServerTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewSqlServerClient(cred, c)
	if err != nil {
		return
	}
	repo = &SqlServerTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
