package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/waf/v20180125"
)

type WafTcInstance struct {
	baseTcInstance
	meta *sdk.DomainInfo
}

func (ins *WafTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewWafTcInstance(instanceId string, meta *sdk.DomainInfo) (ins *WafTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &WafTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
