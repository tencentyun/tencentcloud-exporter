package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cmq/v20190304"
)

type CMQTopicTcInstance struct {
	baseTcInstance
	meta *sdk.TopicSet
}

func (ins *CMQTopicTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewCMQTopicTcInstance(instanceId string, meta *sdk.TopicSet) (ins *CMQTopicTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &CMQTopicTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
