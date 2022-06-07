package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"
)

type ZookeeperTcInstance struct {
	baseTcInstance
	meta *sdk.SREInstance
}

func (ins *ZookeeperTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewZookeeperTcInstance(instanceId string, meta *sdk.SREInstance) (ins *ZookeeperTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &ZookeeperTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
