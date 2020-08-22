package instance

import (
	"fmt"
	"reflect"
	"strconv"
)

// 不支持实例纬度自动查询的namespace
var NotSupportInstances = []string{
	"QCE/COS",
	"QCE/CDN",
}

type TcInstance interface {
	GetInstanceId() string

	GetMonitorQueryKey() string

	// 根据字段名称获取该字段的值, 由各个产品接口具体实现
	GetFieldValueByName(string) (string, error)

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

func (ins *baseTcInstance) GetFieldValueByName(name string) (string, error) {
	v := ins.value.FieldByName(name)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	switch k := v.Kind(); k {
	case reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.String:
		return v.String(), nil
	default:
		return "", fmt.Errorf("value type not support, type=%s", k)
	}
}
