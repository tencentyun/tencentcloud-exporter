package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sqlserver/v20180328"
)

type SqlServerTcInstance struct {
	baseTcInstance
	meta *sdk.DBInstance
}

func (ins *SqlServerTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewSqlServerTcInstance(instanceId string, meta *sdk.DBInstance) (ins *SqlServerTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &SqlServerTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
