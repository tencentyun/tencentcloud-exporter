package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"
)

type TdmqTcInstance struct {
	baseTcInstance
	meta *sdk.AMQPClusterDetail
}

func (ins *TdmqTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewTdmqTcInstance(instanceId string, meta *sdk.AMQPClusterDetail) (ins *TdmqTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &TdmqTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
