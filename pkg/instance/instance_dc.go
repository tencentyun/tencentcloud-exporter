package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
)

type DcTcInstance struct {
	baseTcInstance
	meta *sdk.DirectConnect
}

func (ins *DcTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewDcTcInstance(instanceId string, meta *sdk.DirectConnect) (ins *DcTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &DcTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
