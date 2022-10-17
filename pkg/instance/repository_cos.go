package instance

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sdk "github.com/tencentyun/cos-go-sdk-v5"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/COS", NewCosTcInstanceRepository)
}

type CosTcInstanceRepository struct {
	region string
	client *sdk.Client
	logger log.Logger
}

func (repo *CosTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *CosTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	resp, _, err := repo.client.Service.Get(context.Background())
	if err != nil {
		return
	}

	if len(resp.Buckets) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Buckets[0]
	instance, err = NewCosTcInstance(id, &meta)
	if err != nil {
		return
	}
	return
}

func (repo *CosTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *CosTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	// req := sdk.NewDescribeVpnConnectionsRequest()
	// var offset uint64 = 0
	// var limit uint64 = 100
	// var total int64 = -1
	//
	// req.Offset = &offset
	// req.Limit = &limit

	// getMoreInstances:
	resp, _, err := repo.client.Service.Get(context.Background())
	if err != nil {
		return
	}
	for _, meta := range resp.Buckets {
		// when region is ap-guangzhou, will get all buckets in every region.
		// need to filter by region
		if meta.Region != repo.region {
			continue
		}
		ins, e := NewCosTcInstance(meta.Name, &meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create Cos instance fail", "id", meta.Name)
			continue
		}
		instances = append(instances, ins)
	}
	// offset += limit
	// if offset < uint64(total) {
	// 	req.Offset = &offset
	// 	goto getMoreInstances
	// }

	return
}

func NewCosTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewCosClient(cred, c)
	if err != nil {
		return
	}
	repo = &CosTcInstanceRepository{
		region: c.Credential.Region,
		client: cli,
		logger: logger,
	}
	return
}
