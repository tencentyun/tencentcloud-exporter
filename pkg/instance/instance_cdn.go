package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
)

type CdnTcInstance struct {
	baseTcInstance
	meta *sdk.BriefDomain
}

func (ins *CdnTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewCdnTcInstance(instanceId string, meta *sdk.BriefDomain) (ins *CdnTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &CdnTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
