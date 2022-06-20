package instance

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/VPNX", NewVpnxTcInstanceRepository)
}

type VpnxTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *VpnxTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *VpnxTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeVpnConnectionsRequest()
	// req.Filters.Name = []*string{&id}
	resp, err := repo.client.DescribeVpnConnections(req)
	if err != nil {
		return
	}
	if len(resp.Response.VpnConnectionSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.VpnConnectionSet[0]
	instance, err = NewVpnxTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *VpnxTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *VpnxTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeVpnConnectionsRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeVpnConnections(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.VpnConnectionSet {
		ins, e := NewVpnxTcInstance(*meta.VpnConnectionId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create vpnx instance fail", "id", *meta.VpnConnectionId)
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

func NewVpnxTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewVpvClient(cred, c)
	if err != nil {
		return
	}
	repo = &VpnxTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
