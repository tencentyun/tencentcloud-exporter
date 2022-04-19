package instance

import (
	"fmt"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dcdb/v20180411"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/TDMYSQL", NewDcdbTcInstanceRepository)
}

type DcdbTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *DcdbTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *DcdbTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDCDBInstancesRequest()
	req.InstanceIds = []*string{&id}
	repo.credential.Refresh()
	resp, err := repo.client.DescribeDCDBInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.Instances) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.Instances[0]
	instance, err = NewDcdbTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *DcdbTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *DcdbTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDCDBInstancesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	repo.credential.Refresh()
	resp, err := repo.client.DescribeDCDBInstances(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.Instances {
		ins, e := NewDcdbTcInstance(*meta.InstanceId, meta)
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

func NewDcdbTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewDCDBClient(cred, c)
	if err != nil {
		return
	}
	repo = &DcdbTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
