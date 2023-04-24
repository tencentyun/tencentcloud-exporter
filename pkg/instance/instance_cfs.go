package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cfs/v20190719"
)

type CfsTcInstance struct {
	baseTcInstance
	meta *sdk.FileSystemInfo
}

func (ins *CfsTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewCfsTcInstance(instanceId string, meta *sdk.FileSystemInfo) (ins *CfsTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &CfsTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
