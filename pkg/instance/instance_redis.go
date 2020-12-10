package instance

import (
	"fmt"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	"reflect"
)

type RedisTcInstance struct {
	baseTcInstance
	meta *sdk.InstanceSet
}

func NewRedisTcInstance(instanceId string, meta *sdk.InstanceSet) (ins *RedisTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &RedisTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}

func (ins *RedisTcInstance) GetMeta() interface{} {
	return ins.meta
}
