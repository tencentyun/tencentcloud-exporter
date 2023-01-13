package instance

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/waf/v20180125"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/WAF", NewWafTcInstanceRepository)
}

type WafTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *WafTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

var domain = "Domain"
var exactMatchTrue = true

func (repo *WafTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDomainsRequest()
	var offset uint64 = 1
	var limit uint64 = 100

	req.Offset = &offset
	req.Limit = &limit
	req.Filters = []*sdk.FiltersItemNew{{
		Name:       &domain,
		Values:     []*string{&id},
		ExactMatch: &exactMatchTrue,
	}}
	resp, err := repo.client.DescribeDomains(req)
	if err != nil {
		return
	}
	if len(resp.Response.Domains) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.Domains[0]
	instance, err = NewWafTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *WafTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *WafTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDomainsRequest()
	var offset uint64 = 1
	var limit uint64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeDomains(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.Total)
	}
	for _, meta := range resp.Response.Domains {
		ins, e := NewWafTcInstance(*meta.Domain, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create Waf instance fail", "id", *meta.InstanceId)
			continue
		}
		instances = append(instances, ins)
	}
	offset++
	if (offset-1)*limit < uint64(total) {
		req.Offset = &offset
		goto getMoreInstances
	}

	return
}

func NewWafTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewWafClient(cred, c)
	if err != nil {
		return
	}
	repo = &WafTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
