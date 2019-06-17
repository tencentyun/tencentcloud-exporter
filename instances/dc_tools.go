package instances

import (
	dc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
)

type dcToolStruct struct{}

func (me *dcToolStruct) fillStringFilter(filterName string, valueInterface interface{}, request interface{}) (has bool) {

	if valueInterface == nil || request == nil {
		return
	}

	dcxRequest, isDcx := request.(*dc.DescribeDirectConnectTunnelsRequest)

	dcRequest, isDc := request.(*dc.DescribeDirectConnectsRequest)

	str, ok := valueInterface.(string)
	if !ok {
		return
	}
	has = true
	filter := &dc.Filter{}
	filter.Name = &filterName

	filter.Values = []*string{&str}
	if isDcx {
		dcxRequest.Filters = append(dcxRequest.Filters, filter)
	}
	if isDc {
		dcRequest.Filters = append(dcRequest.Filters, filter)
	}

	return
}
