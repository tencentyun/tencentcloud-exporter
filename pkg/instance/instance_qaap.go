package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap/v20180529"
)

type QaapTcInstance struct {
	baseTcInstance
	meta *sdk.ProxyInfo
}

func (ins *QaapTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewQaapTcInstance(instanceId string, meta *sdk.ProxyInfo) (ins *QaapTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &QaapTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
