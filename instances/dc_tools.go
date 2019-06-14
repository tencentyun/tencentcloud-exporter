package instances

import (
	dc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
)

type dcToolStruct struct{}

func (me *dcToolStruct) fillStringFilter(filterName string, valueInterface interface{}, request *dc.DescribeDirectConnectTunnelsRequest) (has bool) {
	if valueInterface == nil || request == nil {
		return
	}
	str, ok := valueInterface.(string)
	if !ok {
		return
	}
	has = true
	filter := &dc.Filter{}
	filter.Name = &filterName
	filter.Values = []*string{&str}
	request.Filters = append(request.Filters, filter)
	return
}
