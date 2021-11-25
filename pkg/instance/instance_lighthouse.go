package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

type LighthouseTcInstance struct {
	baseTcInstance
	meta *sdk.Instance
}

func (ins *LighthouseTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewLighthouseTcInstance(instanceId string, meta *sdk.Instance) (ins *LighthouseTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &LighthouseTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
