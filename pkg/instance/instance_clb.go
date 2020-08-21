package instance

import (
	"fmt"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"reflect"
)

type ClbInstance struct {
	baseTcInstance
	meta *sdk.LoadBalancer
}

func (ins *ClbInstance) GetMonitorQueryKey() string {
	if len(ins.meta.LoadBalancerVips) == 1 {
		return *ins.meta.LoadBalancerVips[0]
	} else {
		return ""
	}
}

func (ins *ClbInstance) GetMeta() interface{} {
	return ins.meta
}

func NewClbTcInstance(instanceId string, meta *sdk.LoadBalancer) (ins *ClbInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &ClbInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
