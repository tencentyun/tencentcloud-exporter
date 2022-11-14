package instance

import (
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"

	selfcommon "github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dts/v20180330"
	dtsNew "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dts/v20211206"

	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
)

func init() {
	registerRepository("QCE/DTS", NewDTSTcInstanceRepository)
}

type DTSTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *DTSTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *DTSTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeSubscribesRequest()
	req.InstanceId = &id
	resp, err := repo.client.DescribeSubscribes(req)
	if err != nil {
		return
	}
	if len(resp.Response.Items) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.Items[0]
	instance, err = NewDtsTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *DTSTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *DTSTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeSubscribesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeSubscribes(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.Items {
		ins, e := NewDtsTcInstance(*meta.SubscribeId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create Dts instance fail", "id", *meta.SubscribeId)
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

// Replications
type DtsTcInstanceReplicationsRepository interface {
	GetReplicationsInfo(instanceId string) (*dtsNew.DescribeSyncJobsResponse, error)
}

type DtsTcInstanceReplicationsRepositoryImpl struct {
	client *dtsNew.Client
	logger log.Logger
}

func (repo *DtsTcInstanceReplicationsRepositoryImpl) GetReplicationsInfo(instanceId string) (*dtsNew.DescribeSyncJobsResponse, error) {
	req := dtsNew.NewDescribeSyncJobsRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	req.Limit = &limit
	req.Offset = &offset
	req.JobId = common.StringPtr(instanceId)
	return repo.client.DescribeSyncJobs(req)
}

func NewDtsTcInstanceReplicationsRepository(cred selfcommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (DtsTcInstanceReplicationsRepository, error) {
	cli, err := client.NewDTSNewClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &DtsTcInstanceReplicationsRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

// MigrateInfos
type DtsTcInstanceMigrateInfosRepository interface {
	GetMigrateInfos(instanceId string) (*sdk.DescribeMigrateJobsResponse, error)
}

type DtsTcInstanceMigrateInfosRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *DtsTcInstanceMigrateInfosRepositoryImpl) GetMigrateInfos(instanceId string) (*sdk.DescribeMigrateJobsResponse, error) {
	req := sdk.NewDescribeMigrateJobsRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	req.Limit = &limit
	req.Offset = &offset
	req.JobId = common.StringPtr(instanceId)
	return repo.client.DescribeMigrateJobs(req)
}

func NewDtsTcInstanceMigrateInfosRepository(cred selfcommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (DtsTcInstanceMigrateInfosRepository, error) {
	cli, err := client.NewDTSClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &DtsTcInstanceMigrateInfosRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

func NewDTSTcInstanceRepository(cred selfcommon.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewDTSClient(cred, c)
	if err != nil {
		return
	}
	repo = &DTSTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
