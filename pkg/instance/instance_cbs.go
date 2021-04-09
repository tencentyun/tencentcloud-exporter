package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
)

type CbsTcInstance struct {
	baseTcInstance
	meta *sdk.Disk
}

func (ins *CbsTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewCbsTcInstance(instanceId string, meta *sdk.Disk) (ins *CbsTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &CbsTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
