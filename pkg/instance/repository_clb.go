package instance

import (
	"fmt"
	"net"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func init() {
	// LB_PUBLIC、LOADBALANCE实例对象是一样的
	registerRepository("QCE/LB_PUBLIC", NewClbTcInstanceRepository)
	registerRepository("QCE/LOADBALANCE", NewClbTcInstanceRepository)
}

var open = "OPEN"

type ClbTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *ClbTcInstanceRepository) GetInstanceKey() string {
	return "LoadBalancerVip"
}

func (repo *ClbTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeLoadBalancersRequest()

	ip := net.ParseIP(id)
	if ip != nil {
		ipstr := ip.String()
		req.LoadBalancerVips = []*string{&ipstr}
	} else {
		req.LoadBalancerIds = []*string{&id}
	}
	req.LoadBalancerType = &open
	resp, err := repo.client.DescribeLoadBalancers(req)
	if err != nil {
		return
	}

	if len(resp.Response.LoadBalancerSet) == 0 {
		return nil, fmt.Errorf("loadBalancer instance not found")
	} else if len(resp.Response.LoadBalancerSet) > 1 {
		return nil, fmt.Errorf("response instanceDetails size != 1")
	}
	meta := resp.Response.LoadBalancerSet[0]
	instance, err = NewClbTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *ClbTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *ClbTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeLoadBalancersRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit
	req.LoadBalancerType = &open

getMoreInstances:
	resp, err := repo.client.DescribeLoadBalancers(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.LoadBalancerSet {
		ins, e := NewClbTcInstance(*meta.LoadBalancerId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create clb instance fail", "id", *meta.LoadBalancerId)
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

func NewClbTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewClbClient(cred, c)
	if err != nil {
		return
	}
	repo = &ClbTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
