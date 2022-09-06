package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type VbcTcInstance struct {
	baseTcInstance
	meta *sdk.CCN
}

func (ins *VbcTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewVbcTcInstance(instanceId string, meta *sdk.CCN) (ins *VbcTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &VbcTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
