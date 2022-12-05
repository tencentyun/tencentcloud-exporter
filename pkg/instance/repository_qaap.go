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

// DescribeProxyGroupList
type QaapTcInstanceInfoRepository interface {
	GetProxyGroupList(instanceId string) (*sdk.DescribeProxyGroupListResponse, error)
	GetUDPListenersInfo(instanceId string) (*sdk.DescribeUDPListenersResponse, error)
	GetTCPListenersInfo(instanceId string) (*sdk.DescribeTCPListenersResponse, error)
}

type QaapTcInstanceInfoRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *QaapTcInstanceInfoRepositoryImpl) GetProxyGroupList(instanceId string) (*sdk.DescribeProxyGroupListResponse, error) {
	req := sdk.NewDescribeProxyGroupListRequest()
	var offset int64 = 0
	var limit int64 = 100
	var projectId int64 = -1
	req.Limit = &limit
	req.Offset = &offset
	req.ProjectId = &projectId
	return repo.client.DescribeProxyGroupList(req)
}
func (repo *QaapTcInstanceInfoRepositoryImpl) GetUDPListenersInfo(instanceId string) (*sdk.DescribeUDPListenersResponse, error) {
	req := sdk.NewDescribeUDPListenersRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	req.Limit = &limit
	req.Offset = &offset
	req.ProxyId = selfcommon.StringPtr(instanceId)
	return repo.client.DescribeUDPListeners(req)
}
func (repo *QaapTcInstanceInfoRepositoryImpl) GetTCPListenersInfo(instanceId string) (*sdk.DescribeTCPListenersResponse, error) {
	req := sdk.NewDescribeTCPListenersRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	req.Limit = &limit
	req.Offset = &offset
	req.ProxyId = selfcommon.StringPtr(instanceId)
	return repo.client.DescribeTCPListeners(req)
}

func NewQaapTcInstanceInfoRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (QaapTcInstanceInfoRepository, error) {
	cli, err := client.NewGAAPClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &QaapTcInstanceInfoRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

// 内部接口
type ProxyInstancesResponse struct {
	TotalCount int64
	ProxySet   []ProxyDetail
}
type ProxyInstancesRsp struct {
	Response ProxyInstancesResponse
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
	Response NoneBgpIpListResponse
}
type NoneBgpIpListResponse struct {
	TotalCount  int64
	InstanceSet []InstanceDetail
}
type InstanceDetail struct {
	IP      string
	Isp     string
	ProxyId string
	GroupId string
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
		"Limit":    100,
		"Offset":   0,
		"ProxyIds": []string{instanceId},
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
	json.Unmarshal(response.GetBody(), &proxyInstancesRsp)
	return proxyInstancesRsp, nil
}

func (repo *CommonQaapTcInstanceRepositoryImpl) GetCommonQaapNoneBgpIpList(instanceId string) (NoneBgpIpListRsp, error) {
	var noneBgpIpListRsp NoneBgpIpListRsp
	request := tchttp.NewCommonRequest("gaap", "2018-05-29", "DescribeNoneBgpIpList")
	body := map[string]interface{}{
		"Limit":  100,
		"Offset": 0,
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
