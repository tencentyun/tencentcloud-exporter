package instance

import (
	"fmt"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/LB", NewEIPTcInstanceRepository)
}

type EIPTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *EIPTcInstanceRepository) GetInstanceKey() string {
	return "eip"
}

func (repo *EIPTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeAddressesRequest()
	req.AddressIds = []*string{&id}
	resp, err := repo.client.DescribeAddresses(req)
	if err != nil {
		return
	}

	var meta *sdk.Address
	if len(resp.Response.AddressSet) == 0 {
		reqV6 := sdk.NewDescribeIp6AddressesRequest()
		reqV6.Ip6AddressIds = []*string{&id}
		respV6, err := repo.client.DescribeIp6Addresses(reqV6)
		if err != nil {
			return nil, err
		}
		if len(respV6.Response.AddressSet) == 1 {
			meta = respV6.Response.AddressSet[0]
		} else {
			return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
		}
	} else if len(resp.Response.AddressSet) == 1 {
		meta = resp.Response.AddressSet[0]
	} else {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	instance, err = NewEIPTcInstance(*meta.AddressIp, meta)
	if err != nil {
		return
	}
	return
}

func (repo *EIPTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *EIPTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeAddressesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeAddresses(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.AddressSet {
		ins, e := NewEIPTcInstance(*meta.AddressIp, meta)
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

	reqV6 := sdk.NewDescribeIp6AddressesRequest()
	offset, limit, total = 0, 100, -1
	reqV6.Offset = &offset
	reqV6.Limit = &limit
getMoreV6Instances:
	respV6, err := repo.client.DescribeIp6Addresses(reqV6)
	if err != nil {
		return
	}
	if total == -1 {
		total = *respV6.Response.TotalCount
	}
	for _, meta := range respV6.Response.AddressSet {
		ins, e := NewEIPTcInstance(*meta.AddressIp, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create cdb instance fail", "id", *meta.InstanceId)
			continue
		}
		instances = append(instances, ins)
	}
	offset += limit
	if offset < total {
		req.Offset = &offset
		goto getMoreV6Instances
	}

	return
}

func NewEIPTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewVpvClient(cred, c)
	if err != nil {
		return
	}
	repo = &EIPTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
