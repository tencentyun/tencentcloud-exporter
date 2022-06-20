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
	registerRepository("QCE/VPNGW", NewVpngwTcInstanceRepository)
}

type VpngwTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *VpngwTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *VpngwTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeVpnGatewaysRequest()
	// req.Filters.Name = []*string{&id}
	resp, err := repo.client.DescribeVpnGateways(req)
	if err != nil {
		return
	}
	if len(resp.Response.VpnGatewaySet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.VpnGatewaySet[0]
	instance, err = NewVpngwTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *VpngwTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *VpngwTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeVpnGatewaysRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeVpnGateways(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.VpnGatewaySet {
		ins, e := NewVpngwTcInstance(*meta.VpnGatewayId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create vpngw instance fail", "id", *meta.VpnGatewayId)
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

func NewVpngwTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewVpvClient(c)
	if err != nil {
		return
	}
	repo = &VpngwTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
