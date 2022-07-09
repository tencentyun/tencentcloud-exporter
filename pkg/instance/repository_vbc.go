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
	registerRepository("QCE/VBC", NewVbcTcInstanceRepository)
}

type VbcTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *VbcTcInstanceRepository) GetInstanceKey() string {
	return "CcnId"
}

func (repo *VbcTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeCcnsRequest()
	req.CcnIds = []*string{&id}
	resp, err := repo.client.DescribeCcns(req)
	if err != nil {
		return
	}
	if len(resp.Response.CcnSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.CcnSet[0]
	instance, err = NewVbcTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *VbcTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *VbcTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeCcnsRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1
	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeCcns(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.CcnSet {
		ins, e := NewVbcTcInstance(*meta.CcnId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create vbc instance fail", "id", *meta.CcnId)
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

func NewVbcTcInstanceRepository(c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewVpvClient(c)
	if err != nil {
		return
	}
	repo = &VbcTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
