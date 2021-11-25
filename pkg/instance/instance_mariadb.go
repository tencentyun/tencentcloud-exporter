package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mariadb/v20170312"
)

type MariaDBTcInstance struct {
	baseTcInstance
	meta *sdk.DBInstance
}

func (ins *MariaDBTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewMariaDBTcInstance(instanceId string, meta *sdk.DBInstance) (ins *MariaDBTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &MariaDBTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
