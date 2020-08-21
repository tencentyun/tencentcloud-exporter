package instance

import (
	"fmt"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	"reflect"
)

type CdbTcInstance struct {
	baseTcInstance
	meta *sdk.InstanceInfo
}

func (ins *CdbTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewCdbTcInstance(instanceId string, meta *sdk.InstanceInfo) (ins *CdbTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &CdbTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
