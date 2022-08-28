package instance

import (
	"fmt"
	"reflect"
	"regexp"
	"unicode"
)

// 每个产品的实例对象, 可用于配置导出指标的额外label填充, 根据字段名获取值
type TcInstance interface {
	// 获取实例的id
	GetInstanceId() string

	// 用于查询云监控数据的主键字段, 一般是实例id
	GetMonitorQueryKey() string

	// 根据字段名称获取该字段的值, 由各个产品接口具体实现
	GetFieldValueByName(string) (string, error)

	// 根据字段名称获取该字段的值, 由各个产品接口具体实现
	GetFieldValuesByName(string) (map[string][]string, error)

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
			// nothing ignore err
		}
	}()
	v := ins.value.FieldByName(name)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	return fmt.Sprintf("%v", v.Interface()), nil
}

func (ins *baseTcInstance) GetFieldValuesByName(name string) (val map[string][]string, err error) {
	defer func() {
		if err := recover(); err != nil {
			// nothing ignore err
		}
	}()
	v := ins.value.FieldByName(name)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	valueMap := make(map[string][]string)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			if v.Index(i).Elem().Kind() == reflect.String {
				valueMap[name] = append(val[name], fmt.Sprintf("%v", v.Index(i).Elem().Interface()))
			} else if v.Index(i).Elem().Kind() == reflect.Struct {
				var tagKey, tagValue reflect.Value

				if v.Index(i).Elem().FieldByName("TagKey").IsValid() && v.Index(i).Elem().FieldByName("TagValue").IsValid() {
					tagKey = v.Index(i).Elem().FieldByName("TagKey")
					tagValue = v.Index(i).Elem().FieldByName("TagValue")
				} else if v.Index(i).Elem().FieldByName("Key").IsValid() && v.Index(i).Elem().FieldByName("Value").IsValid() {
					tagKey = v.Index(i).Elem().FieldByName("Key")
					tagValue = v.Index(i).Elem().FieldByName("Value")
				}
				if tagKey.Kind() == reflect.Ptr {
					tagKey = reflect.Indirect(tagKey)
				}
				if tagValue.Kind() == reflect.Ptr {
					tagValue = reflect.Indirect(tagValue)
				}
				if IsValidTagKey(tagKey.String()) {
					valueMap[tagKey.String()] = append(val[tagKey.String()], tagValue.String())
				}
			}
		}
	} else {
		valueMap[name] = append(val[name], fmt.Sprintf("%v", v.Interface()))
	}
	return valueMap, nil
}

func IsValidTagKey(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
			return false
		}
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_]+$`).MatchString(str) {
		return false
	}
	return true
}
