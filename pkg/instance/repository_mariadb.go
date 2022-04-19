package instance

import (
	"fmt"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mariadb/v20170312"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/MARIADB", NewMariaDBTcInstanceRepository)
}

type MariaDBTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *MariaDBTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *MariaDBTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	req.InstanceIds = []*string{&id}
	repo.credential.Refresh()
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.Instances) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.Instances[0]
	instance, err = NewMariaDBTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *MariaDBTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *MariaDBTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDBInstancesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = 0

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	repo.credential.Refresh()
	resp, err := repo.client.DescribeDBInstances(req)
	if err != nil {
		return
	}
	if total == 0 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.Instances {
		ins, e := NewMariaDBTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create mariadb instance fail", "id", *meta.InstanceId)
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

func NewMariaDBTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewMariaDBClient(cred, c)
	if err != nil {
		return
	}
	repo = &MariaDBTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
