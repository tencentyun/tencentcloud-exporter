package instances

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/common/log"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

/*what we call this product in prom*/
const CvmProductName = "cvm"

type cvmToolStruct struct{}

var cvmTools cvmToolStruct

func init() {
	funcGets[CvmProductName] = getCvmInstancesIds
}

var cvmApiFilterStrs = []string{"Zone", "VpcId", "InstanceName", "InstanceId", "PrivateIpAddress", "PublicIpAddress", "SubnetId", "InstanceChargeType"}
var cvmApiFilterInts = []string{"ProjectId"}

func getCvmInstancesIds(filters map[string]interface{}) (instanceIdsMap map[string]map[string]interface{},
	errRet error) {

	if credentialConfig.AccessKey == "" {
		errRet = fmt.Errorf("error,instantces  client is not  initialized yet")
		log.Errorf(errRet.Error())
		return
	}

	cacheKey := getCacheKey(CvmProductName, filters)

	if instanceIdsMap = getCache(cacheKey, true); instanceIdsMap != nil {
		log.Debugf("product [%s] list from new cache", CvmProductName)
		return
	}

	/*if product api error, we can get from cache.*/
	defer func() {
		if errRet != nil {
			if oldInstanceIdsMap := getCache(cacheKey, false); oldInstanceIdsMap != nil {
				instanceIdsMap = oldInstanceIdsMap
				log.Warnf("product [%s]  from old cache, because product list api error", MysqlProductName)
			}
		}
	}()

	credential := common.NewCredential(
		credentialConfig.AccessKey,
		credentialConfig.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 10

	client, err := cvm.NewClient(credential, credentialConfig.Region, cpf)
	if err != nil {
		log.Error(err.Error())
		errRet = err
		return
	}
	request := cvm.NewDescribeInstancesRequest()

	var apiCanFilters = map[string]bool{}

	for _, v := range cvmApiFilterStrs {
		cvmTools.fillStringFilter(cvmTools.ToMiddlelineLower(v), filters[v], request)
		apiCanFilters[v] = true
	}
	for _, v := range cvmApiFilterInts {
		cvmTools.fillIntFilter(cvmTools.ToMiddlelineLower(v), filters[v], request)
		apiCanFilters[v] = true
	}
	instanceIdsMap = make(map[string]map[string]interface{})
	hasGet := make(map[string]bool)

	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

getMoreInstanceId:
	request.Offset = &offset
	request.Limit = &limit
	response, err := client.DescribeInstances(request)
	if err != nil {
		response, err = client.DescribeInstances(request)
	}

	if err != nil {
		errRet = err
		log.Errorf("api[%s] fail, request body [%s], reason[%s]", request.GetAction(), request.ToJsonString(), errRet.Error())
		return
	}
	if total == -1 {
		total = *response.Response.TotalCount
	}
	if len(response.Response.InstanceSet) == 0 {
		goto hasGetAll
	}
	for _, v := range response.Response.InstanceSet {
		if _, ok := hasGet[*v.InstanceId]; ok {
			errRet = fmt.Errorf("api[%s] return error, has repeat instance id [%s]", request.GetAction(), *v.InstanceId)
			log.Errorf(errRet.Error())
			return
		}
		js, err := json.Marshal(v)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json encode [%s]", request.GetAction(), *v.InstanceId)
			log.Errorf(errRet.Error())
		}
		var data map[string]interface{}

		err = json.Unmarshal(js, &data)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json decode [%s]", request.GetAction(), *v.InstanceId)
			log.Errorf(errRet.Error())
		}
		/*
			Need to convert some heterogeneous data
		*/
		if v.Placement != nil {
			setNonStandardStr("Zone", v.Placement.Zone, data)
			setNonStandardInt64("ProjectId", v.Placement.ProjectId, data)
		}

		var privateIpAddress string
		for _, v := range v.PrivateIpAddresses {
			privateIpAddress = privateIpAddress + "," + (*v)
			data["PrivateIpAddress"] = privateIpAddress
		}
		var publicIpAddress string
		for _, v := range v.PublicIpAddresses {
			publicIpAddress = publicIpAddress + "," + (*v)
			data["PublicIpAddress"] = publicIpAddress
		}
		if v.VirtualPrivateCloud != nil {
			setNonStandardStr("VpcId", v.VirtualPrivateCloud.VpcId, data)
			setNonStandardStr("SubnetId", v.VirtualPrivateCloud.SubnetId, data)
		}
		if meetConditions(filters, data, apiCanFilters) {
			instanceIdsMap[*v.InstanceId] = data
		}
		hasGet[*v.InstanceId] = true
	}
	offset += limit
	if total != -1 && int64(offset) >= total {
		goto hasGetAll
	}
	goto getMoreInstanceId

hasGetAll:
	if len(instanceIdsMap) > 0 {
		setCache(cacheKey, instanceIdsMap)
	}
	return
}
