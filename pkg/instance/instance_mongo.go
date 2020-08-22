package instance

import (
	"fmt"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mongodb/v20190725"
	"reflect"
)

type MongoTcInstance struct {
	baseTcInstance
	meta *sdk.InstanceDetail
}

func (ins *MongoTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewMongoTcInstance(instanceId string, meta *sdk.InstanceDetail) (ins *MongoTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &MongoTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
