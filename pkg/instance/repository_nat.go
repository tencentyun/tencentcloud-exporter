package instance

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/NAT_GATEWAY", NewNatTcInstanceRepository)
}

type NatTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *NatTcInstanceRepository) GetInstanceKey() string {
	return "instanceid"
}

func (repo *NatTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeNatGatewaysRequest()
	req.NatGatewayIds = []*string{&id}
	resp, err := repo.client.DescribeNatGateways(req)
	if err != nil {
		return
	}
	if len(resp.Response.NatGatewaySet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.NatGatewaySet[0]
	instance, err = NewNatTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *NatTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *NatTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeNatGatewaysRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1
	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeNatGateways(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.NatGatewaySet {
		ins, e := NewNatTcInstance(*meta.NatGatewayId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create redis instance fail", "id", *meta.NatGatewayId)
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

func NewNatTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewVpvClient(c)
	if err != nil {
		return
	}
	repo = &NatTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
