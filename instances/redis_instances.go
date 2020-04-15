package instances

import (
	"encoding/json"
	"fmt"
	"github.com/tencentyun/tencentcloud-exporter/lib/ratelimit"

	"github.com/prometheus/common/log"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
)

/*what we call this product in prom*/
const RedisProductName = "redis"
const ClusterRedisProductName = "cluster_redis"

func init() {
	funcGets[RedisProductName] = getRedisInstancesIds
	funcGets[ClusterRedisProductName] = getRedisInstancesIds
}
func getRedisInstancesIds(filters map[string]interface{}) (instanceIdsMap map[string]map[string]interface{},
	errRet error) {

	if credentialConfig.AccessKey == "" {
		errRet = fmt.Errorf("error,instantces  client is not  initialized yet")
		log.Errorf(errRet.Error())
		return
	}

	cacheKey := getCacheKey(RedisProductName, filters)

	if instanceIdsMap = getCache(cacheKey, true); instanceIdsMap != nil {
		log.Debugf("product [%s] list from new  cache", RedisProductName)
		return
	}

	/*if product api error, we can get from cache.*/
	defer func() {
		if errRet != nil {
			if oldInstanceIdsMap := getCache(cacheKey, false); oldInstanceIdsMap != nil {
				instanceIdsMap = oldInstanceIdsMap
				log.Warnf("product [%s]  from old cache, because product list api error", RedisProductName)
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

	client, err := redis.NewClient(credential, credentialConfig.Region, cpf)
	if err != nil {
		errRet = err
		return
	}
	request := redis.NewDescribeInstancesRequest()

	var apiCanFilters = map[string]bool{}

	fiterName := "ProjectId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(int); ok {
		tempInt64 := int64(temp)
		request.ProjectIds = []*int64{&tempInt64}
	}

	fiterName = "InstanceName"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		request.InstanceName = &temp
	}

	fiterName = "InstanceId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		request.InstanceId = &temp
	}

	fiterName = "VpcId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		request.UniqVpcIds = []*string{&temp}
	}

	fiterName = "SubnetId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		request.UniqSubnetIds = []*string{&temp}
	}

	instanceIdsMap = make(map[string]map[string]interface{})
	hasGet := make(map[string]bool)

	var offset uint64 = 0
	var limit uint64 = 20
	var total int64 = -1

getMoreInstanceId:
	request.Offset = &offset
	request.Limit = &limit
	ratelimit.Check(request.GetAction())
	response, err := client.DescribeInstances(request)
	if err != nil {
		ratelimit.Check(request.GetAction())
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

		setNonStandardStr("VpcId", v.UniqVpcId, data)
		setNonStandardStr("SubnetId", v.UniqSubnetId, data)
		setNonStandardInt64("ProjectId", v.ProjectId, data)
		setNonStandardInt64("Port", v.Port, data)
		setNonStandardInt64("ZoneId", v.ZoneId, data)

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
