package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type DcgTcInstance struct {
	baseTcInstance
	meta *sdk.DirectConnectGateway
}

func (ins *DcgTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewDcgTcInstance(instanceId string, meta *sdk.DirectConnectGateway) (ins *DcgTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &DcgTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
