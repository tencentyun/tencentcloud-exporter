package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/postgres/v20170312"
)

type PGTcInstance struct {
	baseTcInstance
	meta *sdk.DBInstance
}

func (ins *PGTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewPGTcInstance(instanceId string, meta *sdk.DBInstance) (ins *PGTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &PGTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
