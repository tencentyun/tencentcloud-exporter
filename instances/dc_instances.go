package instances

import (
	"encoding/json"
	"fmt"
	"github.com/tencentyun/tencentcloud-exporter/lib/ratelimit"

	"github.com/prometheus/common/log"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
)

/*what we call this product in prom*/
const DcProductName = "dc"
const DcxProductName = "dcx"

var dcTools dcToolStruct

func init() {
	funcGets[DcProductName] = getDcInstancesIds
	funcGets[DcxProductName] = getDcxInstancesIds
}

func getDcInstancesIds(filters map[string]interface{}) (instanceIdsMap map[string]map[string]interface{},
	errRet error) {

	credential := common.NewCredential(
		credentialConfig.AccessKey,
		credentialConfig.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 10

	client, err := dc.NewClient(credential, credentialConfig.Region, cpf)
	if err != nil {
		errRet = err
		return
	}
	request := dc.NewDescribeDirectConnectsRequest()

	var apiCanFilters = map[string]bool{}

	nameFilterMap := map[string]string{
		"DirectConnectName": "direct-connect-name",
		"DirectConnectId":   "direct-connect-id"}

	for name, nameInFilter := range nameFilterMap {
		dcTools.fillStringFilter(nameInFilter, filters[name], request)
		apiCanFilters[name] = true
	}

	instanceIdsMap = make(map[string]map[string]interface{})
	hasGet := make(map[string]bool)

	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

getMoreInstanceId:
	request.Offset = &offset
	request.Limit = &limit
	ratelimit.Check(request.GetAction())
	response, err := client.DescribeDirectConnects(request)
	if err != nil {
		ratelimit.Check(request.GetAction())
		response, err = client.DescribeDirectConnects(request)
	}
	if err != nil {
		errRet = err
		log.Errorf("api[%s] fail, request body [%s], reason[%s]", request.GetAction(), request.ToJsonString(), errRet.Error())
		return
	} else {
		log.Debugf("api[%s] success, request body [%s], response[%s]", request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}
	if total == -1 {
		total = *response.Response.TotalCount
	}
	if len(response.Response.DirectConnectSet) == 0 {
		goto hasGetAll
	}

	for _, v := range response.Response.DirectConnectSet {
		if _, ok := hasGet[*v.DirectConnectId]; ok {
			errRet = fmt.Errorf("api[%s] return error, has repeat instance id [%s]", request.GetAction(), *v.DirectConnectId)
			log.Errorf(errRet.Error())
			return
		}
		js, err := json.Marshal(v)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json encode [%s]", request.GetAction(), *v.DirectConnectId)
			log.Errorf(errRet.Error())
		}
		var data map[string]interface{}

		err = json.Unmarshal(js, &data)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json decode [%s]", request.GetAction(), *v.DirectConnectId)
			log.Errorf(errRet.Error())
		}
		if meetConditions(filters, data, apiCanFilters) {
			instanceIdsMap[*v.DirectConnectId] = data
		}
		hasGet[*v.DirectConnectId] = true
	}
	offset += limit
	if total != -1 && int64(offset) >= total {
		goto hasGetAll
	}
	goto getMoreInstanceId

hasGetAll:
	return
}

func getDcxInstancesIds(filters map[string]interface{}) (instanceIdsMap map[string]map[string]interface{},
	errRet error) {

	if credentialConfig.AccessKey == "" {
		errRet = fmt.Errorf("error,instantces  client is not  initialized yet")
		log.Errorf(errRet.Error())
		return
	}

	credential := common.NewCredential(
		credentialConfig.AccessKey,
		credentialConfig.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 10

	client, err := dc.NewClient(credential, credentialConfig.Region, cpf)
	if err != nil {
		errRet = err
		return
	}
	request := dc.NewDescribeDirectConnectTunnelsRequest()

	var apiCanFilters = map[string]bool{}

	nameFilterMap := map[string]string{
		"DirectConnectTunnelName": "direct-connect-tunnel-name",
		"DirectConnectTunnelId":   "direct-connect-tunnel-id"}

	for name, nameInFilter := range nameFilterMap {
		dcTools.fillStringFilter(nameInFilter, filters[name], request)
		apiCanFilters[name] = true
	}

	instanceIdsMap = make(map[string]map[string]interface{})
	hasGet := make(map[string]bool)

	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

getMoreInstanceId:
	request.Offset = &offset
	request.Limit = &limit
	response, err := client.DescribeDirectConnectTunnels(request)
	if err != nil {
		response, err = client.DescribeDirectConnectTunnels(request)
	}

	if err != nil {
		errRet = err
		log.Errorf("api[%s] fail, request body [%s], reason[%s]", request.GetAction(), request.ToJsonString(), errRet.Error())
		return
	} else {
		log.Debugf("api[%s] success, request body [%s], response[%s]", request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}
	if total == -1 {
		total = *response.Response.TotalCount
	}
	if len(response.Response.DirectConnectTunnelSet) == 0 {
		goto hasGetAll
	}

	for _, v := range response.Response.DirectConnectTunnelSet {
		if _, ok := hasGet[*v.DirectConnectTunnelId]; ok {
			errRet = fmt.Errorf("api[%s] return error, has repeat instance id [%s]", request.GetAction(), *v.DirectConnectTunnelId)
			log.Errorf(errRet.Error())
			return
		}
		js, err := json.Marshal(v)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json encode [%s]", request.GetAction(), *v.DirectConnectTunnelId)
			log.Errorf(errRet.Error())
		}
		var data map[string]interface{}

		err = json.Unmarshal(js, &data)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json decode [%s]", request.GetAction(), *v.DirectConnectTunnelId)
			log.Errorf(errRet.Error())
		}
		if meetConditions(filters, data, apiCanFilters) {
			instanceIdsMap[*v.DirectConnectTunnelId] = data
		}
		hasGet[*v.DirectConnectTunnelId] = true
	}
	offset += limit
	if total != -1 && int64(offset) >= total {
		goto hasGetAll
	}
	goto getMoreInstanceId

hasGetAll:
	return
}
