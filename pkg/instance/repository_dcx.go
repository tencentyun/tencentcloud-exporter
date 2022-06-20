package instance

import (
	"fmt"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/DCX", NewDcxTcInstanceRepository)
}

type DcxTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *DcxTcInstanceRepository) GetInstanceKey() string {
	return "directConnectConnId"
}

func (repo *DcxTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDirectConnectTunnelsRequest()
	req.DirectConnectTunnelIds = []*string{&id}
	resp, err := repo.client.DescribeDirectConnectTunnels(req)
	if err != nil {
		return
	}
	if len(resp.Response.DirectConnectTunnelSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.DirectConnectTunnelSet[0]
	instance, err = NewDcxTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *DcxTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *DcxTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDirectConnectTunnelsRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1
	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeDirectConnectTunnels(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.DirectConnectTunnelSet {
		ins, e := NewDcxTcInstance(*meta.DirectConnectTunnelId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create dcx instance fail", "id", *meta.DirectConnectId)
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

func NewDcxTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewDcClient(cred, c)
	if err != nil {
		return
	}
	repo = &DcxTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
