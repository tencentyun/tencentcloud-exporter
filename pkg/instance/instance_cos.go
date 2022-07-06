package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentyun/cos-go-sdk-v5"
)

type CosTcInstance struct {
	baseTcInstance
	meta *sdk.Bucket
}

func (ins *CosTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewCosTcInstance(instanceId string, meta *sdk.Bucket) (ins *CosTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &CosTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
