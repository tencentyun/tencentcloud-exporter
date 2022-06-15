package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cynosdb/v20190107"
)

type CynosdbTcInstance struct {
	baseTcInstance
	meta *sdk.CynosdbInstance
}

func (ins *CynosdbTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewCynosdbTcInstance(instanceId string, meta *sdk.CynosdbInstance) (ins *CynosdbTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &CynosdbTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
