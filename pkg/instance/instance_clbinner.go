package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

type ClbinnerInstance struct {
	baseTcInstance
	meta *sdk.LoadBalancer
}

func (ins *ClbinnerInstance) GetMonitorQueryKey() string {
	if len(ins.meta.LoadBalancerVips) == 1 {
		return *ins.meta.LoadBalancerVips[0]
	} else {
		return ""
	}
}

func (ins *ClbinnerInstance) GetMeta() interface{} {
	return ins.meta
}

func NewClbinnerTcInstance(instanceId string, meta *sdk.LoadBalancer) (ins *ClbinnerInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &ClbinnerInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
