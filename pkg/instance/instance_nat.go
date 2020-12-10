package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type NatTcInstance struct {
	baseTcInstance
	meta *sdk.NatGateway
}

func (ins *NatTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewNatTcInstance(instanceId string, meta *sdk.NatGateway) (ins *NatTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &NatTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
