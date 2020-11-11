package metric

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
)

type Labels map[string]string

func (l *Labels) Md5() (string, error) {
	h := md5.New()
	jb, err := json.Marshal(l)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(jb)), nil
}

// 代表一个指标的labels
type TcmLabels struct {
	queryLableNames    []string // 用于查询数据的条件标签
	instanceLabelNames []string // 从获取实例对象动态获取字段值的标签
	constLabels        Labels   // 用户自定义的常量标签
	Names              []string // 所有标签名列表
}

// 根据标签名, 获取所有标签的值
func (l *TcmLabels) GetValues(filters map[string]string, ins instance.TcInstance) (values []string, err error) {
	lowerKeyFilters := map[string]string{}
	for k, v := range filters {
		lowerKeyFilters[strings.ToLower(k)] = v
	}

	nameValues := map[string]string{}
	for _, name := range l.queryLableNames {
		v, ok := lowerKeyFilters[strings.ToLower(name)]
		if ok {
			nameValues[name] = v
		} else {
			nameValues[name] = ""
		}
	}
	for _, name := range l.instanceLabelNames {
		v, e := ins.GetFieldValueByName(name)
		if e != nil {
			nameValues[name] = ""
		} else {
			nameValues[name] = v
		}
	}
	for name, value := range l.constLabels {
		nameValues[name] = value
	}
	for _, name := range l.Names {
		values = append(values, nameValues[name])
	}
	return
}

func NewTcmLabels(qln []string, iln []string, cl Labels) (*TcmLabels, error) {
	var labelNames []string
	labelNames = append(labelNames, qln...)
	labelNames = append(labelNames, iln...)
	for lname := range cl {
		labelNames = append(labelNames, lname)
	}
	var uniq = map[string]bool{}
	for _, name := range labelNames {
		uniq[name] = true
	}
	var uniqLabelNames []string
	for n := range uniq {
		uniqLabelNames = append(uniqLabelNames, n)
	}
	sort.Strings(uniqLabelNames)

	l := &TcmLabels{
		queryLableNames:    qln,
		instanceLabelNames: iln,
		constLabels:        cl,
		Names:              uniqLabelNames,
	}
	return l, nil
}
