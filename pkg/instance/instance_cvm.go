package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type CvmTcInstance struct {
	baseTcInstance
	meta *sdk.Instance
}

func (ins *CvmTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewCvmTcInstance(instanceId string, meta *sdk.Instance) (ins *CvmTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &CvmTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
