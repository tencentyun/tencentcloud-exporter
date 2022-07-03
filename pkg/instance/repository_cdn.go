package instance

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/CDN", NewCdnTcInstanceRepository)
}

type CdnTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *CdnTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *CdnTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDomainsRequest()
	// req.Filters.Name = []*string{&id}
	resp, err := repo.client.DescribeDomains(req)
	if err != nil {
		return
	}
	if len(resp.Response.Domains) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.Domains[0]
	instance, err = NewCdnTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *CdnTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *CdnTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDomainsRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeDomains(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalNumber
	}
	for _, meta := range resp.Response.Domains {
		ins, e := NewCdnTcInstance(*meta.Domain, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create Cdn instance fail", "id", *meta.Domain)
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

func NewCdnTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewCdnClient(c)
	if err != nil {
		return
	}
	repo = &CdnTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
