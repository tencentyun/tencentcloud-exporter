package instances

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prometheus/common/log"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

/*what we call this product in prom*/
const ClbProductName = "public_clb"

const LoadBalancerTypeInternet = "OPEN"
const LoadBalancerTypeInternal = "INTERNAL"

func init() {
	funcGets[ClbProductName] = getClbLoadBalancerVips
}

func getClbLoadBalancerVips(filters map[string]interface{}) (instanceIdsMap map[string]map[string]interface{},
	errRet error) {

	if credentialConfig.AccessKey == "" {
		errRet = fmt.Errorf("error,instantces  client is not  initialized yet")
		log.Errorf(errRet.Error())
		return
	}

	cacheKey := getCacheKey(ClbProductName, filters)

	if instanceIdsMap = getCache(cacheKey, true); instanceIdsMap != nil {
		log.Debugf("product [%s] list from new cache", ClbProductName)
		return
	}

	/*if product api error, we can get from cache.*/
	defer func() {
		if errRet != nil {
			if oldInstanceIdsMap := getCache(cacheKey, false); oldInstanceIdsMap != nil {
				instanceIdsMap = oldInstanceIdsMap
				log.Warnf("product [%s]  from old cache, because product list api error", ClbProductName)
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

	client, err := clb.NewClient(credential, credentialConfig.Region, cpf)
	if err != nil {
		errRet = err
		return
	}
	request := clb.NewDescribeLoadBalancersRequest()

	var apiCanFilters = map[string]bool{}

	fiterName := "ProjectId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(int); ok {
		tempInt64 := int64(temp)
		request.ProjectId = &tempInt64
	}
	fiterName = "VpcId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		request.VpcId = &temp
	}

	fiterName = "LoadBalancerName"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		request.LoadBalancerName = &temp
	}

	fiterName = "LoadBalancerVip"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		request.LoadBalancerVips = []*string{&temp}
	}

	instanceIdsMap = make(map[string]map[string]interface{})
	hasGet := make(map[string]bool)

	var offset int64 = 0
	var limit int64 = 20
	var total int64 = -1

getMoreInstanceId:
	request.Offset = &offset
	request.Limit = &limit
	response, err := client.DescribeLoadBalancers(request)
	if err != nil {
		response, err = client.DescribeLoadBalancers(request)
	}
	if err != nil {
		errRet = err
		log.Errorf("api[%s] fail, request body [%s], reason[%s]", request.GetAction(), request.ToJsonString(), errRet.Error())
		return
	}
	if total == -1 {
		total = int64(*response.Response.TotalCount)
	}
	if len(response.Response.LoadBalancerSet) == 0 {
		goto hasGetAll
	}
	for _, v := range response.Response.LoadBalancerSet {
		if _, ok := hasGet[*v.LoadBalancerId]; ok {
			errRet = fmt.Errorf("api[%s] return error, has repeat instance id [%s]", request.GetAction(), *v.LoadBalancerId)
			log.Errorf(errRet.Error())
			return
		}
		js, err := json.Marshal(v)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json encode [%s]", request.GetAction(), *v.LoadBalancerId)
			log.Errorf(errRet.Error())
		}
		var data map[string]interface{}

		err = json.Unmarshal(js, &data)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json decode [%s]", request.GetAction(), *v.LoadBalancerId)
			log.Errorf(errRet.Error())
		}

		if len(v.LoadBalancerVips) == 0 {
			log.Warnf("This clb [%s] has none LoadBalancerVip", *v.LoadBalancerId)
			continue
		}

		if len(v.LoadBalancerVips) != 1 {
			log.Errorf("This clb [%s] has %d LoadBalancerVips,we can not solute now.", *v.LoadBalancerId, len(v.LoadBalancerVips))
			continue
		}
		loadBalancerVip := *(v.LoadBalancerVips[0])

		if strings.ToUpper(*v.LoadBalancerType) != LoadBalancerTypeInternet {
			continue
		}
		if v.ProjectId != nil {
			data["ProjectId"] = int64(*v.ProjectId)
		}
		if meetConditions(filters, data, apiCanFilters) {
			data["LoadBalancerVip"] = loadBalancerVip
			instanceIdsMap[loadBalancerVip] = data
		}
		hasGet[*v.LoadBalancerId] = true
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
