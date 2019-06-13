package instances

import (
	"fmt"

	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

/*
	ProjectId ---- > project-id
*/
func (me *cvmToolStruct) ToMiddlelineLower(s string) string {

	var interval byte = 'a' - 'A'

	b := make([]byte, 0, len(s))

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += interval
			if i != 0 {
				b = append(b, '-')
			}
		}
		b = append(b, c)
	}
	return string(b)
}

func (me *cvmToolStruct) fillStringFilter(filterName string, valueInterface interface{}, request *cvm.DescribeInstancesRequest) (has bool) {
	if valueInterface == nil || request == nil {
		return
	}
	str, ok := valueInterface.(string)
	if !ok {
		return
	}
	has = true
	filter := &cvm.Filter{}
	filter.Name = &filterName
	filter.Values = []*string{&str}
	request.Filters = append(request.Filters, filter)
	return
}

func (me *cvmToolStruct) fillIntFilter(filterName string, valueInterface interface{}, request *cvm.DescribeInstancesRequest) (has bool) {
	if valueInterface == nil || request == nil {
		return
	}
	intv, ok := valueInterface.(int)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d", intv)
	has = true
	filter := &cvm.Filter{}
	filter.Name = &filterName
	filter.Values = []*string{&str}
	request.Filters = append(request.Filters, filter)
	return
}
