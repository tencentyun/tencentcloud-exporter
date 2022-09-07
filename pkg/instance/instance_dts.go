package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dts/v20180330"
)

type DtsTcInstance struct {
	baseTcInstance
	meta            *sdk.SubscribeInfo
	subscribeMeta   *sdk.SubscribeInfo
	migrateInfoMeta *sdk.MigrateJobInfo
}

func (ins *DtsTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewDtsTcInstance(instanceId string, meta *sdk.SubscribeInfo) (ins *DtsTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &DtsTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
