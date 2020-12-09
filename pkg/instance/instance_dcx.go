package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
)

type DcxTcInstance struct {
	baseTcInstance
	meta *sdk.DirectConnectTunnel
}

func (ins *DcxTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewDcxTcInstance(instanceId string, meta *sdk.DirectConnectTunnel) (ins *DcxTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &DcxTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
