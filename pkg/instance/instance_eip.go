package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type EIPInstance struct {
	baseTcInstance
	meta *sdk.Address
}

func (ins *EIPInstance) GetMeta() interface{} {
	return ins.meta
}

func NewEIPTcInstance(instanceId string, meta *sdk.Address) (ins *EIPInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &EIPInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
