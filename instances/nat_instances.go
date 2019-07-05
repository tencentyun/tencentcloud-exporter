package instances

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

/*what we call this product in prom*/
const NatProductName = "nat"

func init() {
	funcGets[NatProductName] = getNatInstancesIds
}

func getNatInstancesIds(filters map[string]interface{}) (instanceIdsMap map[string]map[string]interface{},
	errRet error) {

	if credentialConfig.AccessKey == "" {
		errRet = fmt.Errorf("error,instantces  client is not  initialized yet")
		log.Errorf(errRet.Error())
		return
	}

	cacheKey := getCacheKey(NatProductName, filters)

	if instanceIdsMap = getCache(cacheKey, true); instanceIdsMap != nil {
		log.Debugf("product [%s] list from new cache", NatProductName)
		return
	}

	/*if product api error, we can get from cache.*/
	defer func() {
		if errRet != nil {
			if oldInstanceIdsMap := getCache(cacheKey, false); oldInstanceIdsMap != nil {
				instanceIdsMap = oldInstanceIdsMap
				log.Warnf("product [%s]  from old cache, because product list api error", NatProductName)
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

	client, err := vpc.NewClient(credential, credentialConfig.Region, cpf)
	if err != nil {
		errRet = err
		return
	}
	request := vpc.NewDescribeNatGatewaysRequest()
	request.Filters = make([]*vpc.Filter, 0, 3)

	var apiCanFilters = map[string]bool{}

	fiterName := "VpcId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		name := "vpc-id";
		request.Filters = append(request.Filters,
			&vpc.Filter{
				Name:   &name,
				Values: []*string{&temp},
			})
	}

	fiterName = "InstanceId"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		name := "nat-gateway-id";
		request.Filters = append(request.Filters,
			&vpc.Filter{
				Name:   &name,
				Values: []*string{&temp},
			})
	}

	fiterName = "InstanceName"
	apiCanFilters[fiterName] = true
	if temp, ok := filters[fiterName].(string); ok {
		name := "nat-gateway-name";
		request.Filters = append(request.Filters,
			&vpc.Filter{
				Name:   &name,
				Values: []*string{&temp},
			})
	}

	instanceIdsMap = make(map[string]map[string]interface{})

	hasGet := make(map[string]bool)

	var offset uint64 = 0
	var limit uint64 = 20
	var total int64 = -1

getMoreInstanceId:
	request.Offset = &offset
	request.Limit = &limit
	response, err := client.DescribeNatGateways(request)
	if err != nil {
		response, err = client.DescribeNatGateways(request)
	}
	if err != nil {
		errRet = err
		log.Errorf("api[%s] fail, request body [%s], reason[%s]", request.GetAction(), request.ToJsonString(), errRet.Error())
		return
	}
	if total == -1 {
		total = int64(*response.Response.TotalCount)
	}
	if len(response.Response.NatGatewaySet) == 0 {
		goto hasGetAll
	}
	for _, v := range response.Response.NatGatewaySet {
		if _, ok := hasGet[*v.NatGatewayId]; ok {
			errRet = fmt.Errorf("api[%s] return error, has repeat instance id [%s]", request.GetAction(), *v.NatGatewayId)
			log.Errorf(errRet.Error())
			return
		}
		js, err := json.Marshal(v)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json encode [%s]", request.GetAction(), *v.NatGatewayId)
			log.Errorf(errRet.Error())
		}
		var data map[string]interface{}

		err = json.Unmarshal(js, &data)
		if err != nil {
			errRet = fmt.Errorf("api[%s] return error, can not json decode [%s]", request.GetAction(), *v.NatGatewayId)
			log.Errorf(errRet.Error())
		}

		setNonStandardStr("InstanceId", v.NatGatewayId, data)
		setNonStandardStr("InstanceName", v.NatGatewayName, data)

		if meetConditions(filters, data, apiCanFilters) {
			instanceIdsMap[*v.NatGatewayId] = data
		}
		hasGet[*v.NatGatewayId] = true
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
