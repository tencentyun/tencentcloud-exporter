package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ckafka/v20190819"
)

type kafkaTcInstance struct {
	baseTcInstance
	meta *sdk.Instance
}

func (ins *kafkaTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewKafkaTcInstance(instanceId string, meta *sdk.Instance) (ins *kafkaTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &kafkaTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
