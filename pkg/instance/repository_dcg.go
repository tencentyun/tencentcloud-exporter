package instance

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/DCG", NewDcgTcInstanceRepository)
}

type DcgTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *DcgTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *DcgTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDirectConnectGatewaysRequest()
	// req.Filters.Name = []*string{&id}
	resp, err := repo.client.DescribeDirectConnectGateways(req)
	if err != nil {
		return
	}
	if len(resp.Response.DirectConnectGatewaySet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.DirectConnectGatewaySet[0]
	instance, err = NewDcgTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *DcgTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *DcgTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDirectConnectGatewaysRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeDirectConnectGateways(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.DirectConnectGatewaySet {
		ins, e := NewDcgTcInstance(*meta.DirectConnectGatewayId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create Dcg instance fail", "id", *meta.DirectConnectGatewayId)
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

func NewDcgTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewVpvClient(cred, c)
	if err != nil {
		return
	}
	repo = &DcgTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
