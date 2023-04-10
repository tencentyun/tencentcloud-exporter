package instance

import (
	"fmt"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cfs/v20190719"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/CFS", NewCfsTcInstanceRepository)
}

type CfsTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *CfsTcInstanceRepository) GetInstanceKey() string {
	return "FileSystemId"
}

func (repo *CfsTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeCfsFileSystemsRequest()
	req.FileSystemId = &id
	resp, err := repo.client.DescribeCfsFileSystems(req)
	if err != nil {
		return
	}
	if len(resp.Response.FileSystems) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.FileSystems[0]
	instance, err = NewCfsTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *CfsTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *CfsTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeCfsFileSystemsRequest()

	// getMoreInstances:
	resp, err := repo.client.DescribeCfsFileSystems(req)
	if err != nil {
		return
	}
	for _, meta := range resp.Response.FileSystems {
		ins, e := NewCfsTcInstance(*meta.FileSystemId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create Cfs instance fail", "id", *meta.FileSystemId)
			continue
		}
		instances = append(instances, ins)
	}
	// goto getMoreInstances

	return
}

func NewCfsTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewCfsClient(cred, c)
	if err != nil {
		return
	}
	repo = &CfsTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
