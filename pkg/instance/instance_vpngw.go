package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type VpngwTcInstance struct {
	baseTcInstance
	meta *sdk.VpnGateway
}

func (ins *VpngwTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewVpngwTcInstance(instanceId string, meta *sdk.VpnGateway) (ins *VpngwTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &VpngwTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
