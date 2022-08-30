package instance

import (
	"fmt"
	"reflect"
)

// 每个产品的实例对象, 可用于配置导出指标的额外label填充, 根据字段名获取值
type TcInstance interface {
	// 获取实例的id
	GetInstanceId() string

	// 用于查询云监控数据的主键字段, 一般是实例id
	GetMonitorQueryKey() string

	// 根据字段名称获取该字段的值, 由各个产品接口具体实现
	GetFieldValueByName(string) (string, error)

	// 获取实例raw元数据, 每个实例类型不一样
	GetMeta() interface{}
}

type baseTcInstance struct {
	instanceId string
	value      reflect.Value
}

func (ins *baseTcInstance) GetInstanceId() string {
	return ins.instanceId
}

func (ins *baseTcInstance) GetMonitorQueryKey() string {
	return ins.instanceId
}

func (ins *baseTcInstance) GetFieldValueByName(name string) (val string, err error) {
	defer func() {
		if err := recover(); err != nil {
			//nothing ignore err
		}
	}()
	v := ins.value.FieldByName(name)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	return fmt.Sprintf("%v", v.Interface()), nil
}
