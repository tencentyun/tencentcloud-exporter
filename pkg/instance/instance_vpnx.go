package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type VpnxTcInstance struct {
	baseTcInstance
	meta *sdk.VpnConnection
}

func (ins *VpnxTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewVpnxTcInstance(instanceId string, meta *sdk.VpnConnection) (ins *VpnxTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &VpnxTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
