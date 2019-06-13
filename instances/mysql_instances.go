package instances

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/common/log"
	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

/*what we call this product in prom*/
const MysqlProductName = "mysql"

var test = 0

func init() {
	funcGets[MysqlProductName] = getMysqlInstancesIds
}

func getMysqlInstancesIds(filters map[string]interface{}) (instanceIdsMap map[string]map[string]interface{},
	errRet error) {

	if credentialConfig.AccessKey == "" {
		errRet = fmt.Errorf("error,instantces  client is not  initialized yet")
		log.Errorf(errRet.Error())
		return
	}

	cacheKey := getCacheKey(MysqlProductName, filters)

	if instanceIdsMap = getCache(cacheKey, true); instanceIdsMap != nil {
		log.Debugf("product [%s] list from new cache", MysqlProductName)
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

	mysqlClient, err := cdb.NewClient(credential, credentialConfig.Region, cpf)
	if err != nil {
		errRet = err
		return
	}
	request := cdb.NewDescribeDBInstancesRequest()

	var apiCanFilters = map[string]bool{}

	fiterName := "ProjectId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(int); ok {
		tempInt64 := int64(temp)
		request.ProjectId = &tempInt64
	}

	fiterName = "InstanceName"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		request.InstanceNames = []*string{&temp}
	}

	fiterName = "InstanceId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		request.InstanceIds = []*string{&temp}
	}

	instanceIdsMap = make(map[string]map[string]interface{})
	hasGet := make(map[string]bool)

	var offset uint64 = 0
	var limit uint64 = 2000
	var total int64 = -1

getMoreInstanceId:
	request.Offset = &offset
	request.Limit = &limit
	response, err := mysqlClient.DescribeDBInstances(request)
	if err != nil {
		response, err = mysqlClient.DescribeDBInstances(request)
	}
	if err != nil {
		errRet = err
		log.Errorf("api[%s] fail, request body [%s], reason[%s]", request.GetAction(), request.ToJsonString(), errRet.Error())
		return
	}
	if total == -1 {
		total = *response.Response.TotalCount
	}
	if len(response.Response.Items) == 0 {
		goto hasGetAll
	}
	for _, v := range response.Response.Items {
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

		setNonStandardStr("VpcId", v.UniqVpcId, data)
		setNonStandardStr("SubnetId", v.UniqSubnetId, data)
		setNonStandardInt64("ProjectId", v.ProjectId, data)

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
