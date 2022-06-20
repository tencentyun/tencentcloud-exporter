package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"
)

type TseTcInstance struct {
	baseTcInstance
	meta *sdk.SREInstance
}

func (ins *TseTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewTseTcInstance(instanceId string, meta *sdk.SREInstance) (ins *TseTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &TseTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
