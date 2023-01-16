package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

type ClbPrivateInstance struct {
	baseTcInstance
	meta *sdk.LoadBalancer
}

func (ins *ClbPrivateInstance) GetMonitorQueryKey() string {
	if len(ins.meta.LoadBalancerVips) == 1 {
		return *ins.meta.LoadBalancerVips[0]
	} else if *ins.meta.AddressIPv6 != "" {
		return *ins.meta.AddressIPv6
	} else {
		return ""
	}
}

func (ins *ClbPrivateInstance) GetMeta() interface{} {
	return ins.meta
}

func NewClbPrivateTcInstance(instanceId string, meta *sdk.LoadBalancer) (ins *ClbPrivateInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &ClbPrivateInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
