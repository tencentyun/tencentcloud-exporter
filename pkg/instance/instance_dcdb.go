package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dcdb/v20180411"
)

type DcdbTcInstance struct {
	baseTcInstance
	meta *sdk.DCDBInstanceInfo
}

func (ins *DcdbTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewDcdbTcInstance(instanceId string, meta *sdk.DCDBInstanceInfo) (ins *DcdbTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &DcdbTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
