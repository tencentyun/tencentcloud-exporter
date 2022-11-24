package instance

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
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

// 内部接口

type Rsp struct {
	TotalCount int64
	ProxySet   []ProxyDetail
}
type ProxyInstancesRsp struct {
	Response Rsp
}
type ProxyDetail struct {
	ProxyId       string
	ProxyName     string
	L4ListenerSet []L4ListenerDetail
	L7ListenerSet []L7ListenerDetail
}
type L4ListenerDetail struct {
	ListenerId   string
	ListenerName string
	Protocol     string
	RsSet        []BoundRsDetail
}
type BoundRsDetail struct {
	RsId   string
	RsInfo string
}
type L7ListenerDetail struct {
	ListenerId      string
	ListenerName    string
	ForwardProtocol string
	RuleSet         []RuleDetail
}
type RuleDetail struct {
	RsSet  []BoundRsDetail
	RuleId string
}
type NoneBgpIpListRsp struct {
	Response Rsp
}
type CommonQaapTcInstanceRepository interface {
	GetCommonQaapProxyInstances(instanceId string) (ProxyInstancesRsp, error)
	GetCommonQaapNoneBgpIpList(instanceId string) (NoneBgpIpListRsp, error)
}

type CommonQaapTcInstanceRepositoryImpl struct {
	client *selfcommon.Client
	logger log.Logger
}

func (repo *CommonQaapTcInstanceRepositoryImpl) GetCommonQaapProxyInstances(instanceId string) (ProxyInstancesRsp, error) {
	var proxyInstancesRsp ProxyInstancesRsp
	request := tchttp.NewCommonRequest("gaap", "2018-05-29", "DescribeProxyInstances")
	body := map[string]interface{}{
		"Limit":    1,
		"Offset":   0,
		"ProxyIds": []string{"link-2r1whx05"},
	}
	// 设置action所需的请求数据
	err := request.SetActionParameters(body)
	if err != nil {
		return proxyInstancesRsp, err
	}
	// 创建common response
	response := tchttp.NewCommonResponse()
	// 发送请求
	err = repo.client.Send(request, response)
	if err != nil {
		fmt.Printf("fail to invoke api: %v \n", err)
	}
	// 获取响应结果
	// fmt.Println(string(response.GetBody()))
	json.Unmarshal(response.GetBody(), &proxyInstancesRsp)
	return proxyInstancesRsp, nil
}

func (repo *CommonQaapTcInstanceRepositoryImpl) GetCommonQaapNoneBgpIpList(instanceId string) (NoneBgpIpListRsp, error) {
	var noneBgpIpListRsp NoneBgpIpListRsp
	request := tchttp.NewCommonRequest("gaap", "2018-05-29", "DescribeNoneBgpIpList")
	body := map[string]interface{}{
		"Limit":    1,
		"Offset":   0,
	}
	// 设置action所需的请求数据
	err := request.SetActionParameters(body)
	if err != nil {
		return noneBgpIpListRsp, err
	}
	// 创建common response
	response := tchttp.NewCommonResponse()
	// 发送请求
	err = repo.client.Send(request, response)
	if err != nil {
		fmt.Printf("fail to invoke api: %v \n", err)
	}
	// 获取响应结果
	// fmt.Println(string(response.GetBody()))
	json.Unmarshal(response.GetBody(), &noneBgpIpListRsp)
	return noneBgpIpListRsp, nil
}

func NewCommonQaapTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (CommonQaapTcInstanceRepository, error) {
	cli := client.NewGAAPCommonClient(cred, c)
	repo := &CommonQaapTcInstanceRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}
