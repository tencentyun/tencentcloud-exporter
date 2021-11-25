package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/memcached/v20190318"
)

type MemcachedTcInstance struct {
	baseTcInstance
	meta *sdk.InstanceListInfo
}

func (ins *MemcachedTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewMemcachedTcInstance(instanceId string, meta *sdk.InstanceListInfo) (ins *MemcachedTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &MemcachedTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
