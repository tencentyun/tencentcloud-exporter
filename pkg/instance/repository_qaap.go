package instance

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap/v20180529"

	selfcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/QAAP", NewQaapTcInstanceRepository)
}

type QaapTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *QaapTcInstanceRepository) GetInstanceKey() string {
	return "channelId"
}

func (repo *QaapTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeProxiesRequest()
	req.ProxyIds = []*string{&id}
	resp, err := repo.client.DescribeProxies(req)
	if err != nil {
		return
	}
	if len(resp.Response.ProxySet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.ProxySet[0]
	instance, err = NewQaapTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *QaapTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *QaapTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeProxiesRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeProxies(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.ProxySet {
		ins, e := NewQaapTcInstance(*meta.ProxyId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create Vbc instance fail", "id", *meta.ProxyId)
			continue
		}
		instances = append(instances, ins)
	}
	offset += limit
	if int64(offset) < total {
		req.Offset = &offset
		goto getMoreInstances
	}

	return
}

func NewQaapTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewGAAPClient(cred, c)
	if err != nil {
		return
	}
	repo = &QaapTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}

// TCPListeners
type QaapTcInstanceTCPListenersRepository interface {
	GetTCPListenersInfo(instanceId string) (*sdk.DescribeTCPListenersResponse, error)
}

type QaapTcInstanceTCPListenersRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *QaapTcInstanceTCPListenersRepositoryImpl) GetTCPListenersInfo(instanceId string) (*sdk.DescribeTCPListenersResponse, error) {
	req := sdk.NewDescribeTCPListenersRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	req.Limit = &limit
	req.Offset = &offset
	req.ProxyId = selfcommon.StringPtr(instanceId)
	return repo.client.DescribeTCPListeners(req)
}

func NewQaapTcInstanceTCPListenersRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (QaapTcInstanceTCPListenersRepository, error) {
	cli, err := client.NewGAAPClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &QaapTcInstanceTCPListenersRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

// UDPListeners
type QaapTcInstanceUDPListenersRepository interface {
	GetUDPListenersInfo(instanceId string) (*sdk.DescribeUDPListenersResponse, error)
}

type QaapTcInstanceUDPListenersRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *QaapTcInstanceUDPListenersRepositoryImpl) GetUDPListenersInfo(instanceId string) (*sdk.DescribeUDPListenersResponse, error) {
	req := sdk.NewDescribeUDPListenersRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	req.Limit = &limit
	req.Offset = &offset
	req.ProxyId = selfcommon.StringPtr(instanceId)
	return repo.client.DescribeUDPListeners(req)
}

func NewQaapTcInstanceUDPListenersRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (QaapTcInstanceUDPListenersRepository, error) {
	cli, err := client.NewGAAPClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &QaapTcInstanceUDPListenersRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}
