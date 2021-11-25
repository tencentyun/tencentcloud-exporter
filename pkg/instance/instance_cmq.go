package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cmq/v20190304"
)

type CMQTcInstance struct {
	baseTcInstance
	meta *sdk.QueueSet
}

func (ins *CMQTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewCMQTcInstance(instanceId string, meta *sdk.QueueSet) (ins *CMQTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &CMQTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
