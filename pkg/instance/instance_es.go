package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/es/v20180416"
)

type ESTcInstance struct {
	baseTcInstance
	meta *sdk.InstanceInfo
}

func (ins *ESTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewESTcInstance(instanceId string, meta *sdk.InstanceInfo) (ins *ESTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &ESTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
