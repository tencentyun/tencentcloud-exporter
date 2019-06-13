package monitor

import (
	"fmt"
	"math"
	"time"

	"github.com/tencentyun/tencentcloud-exporter/config"

	"github.com/prometheus/common/log"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	remoteMonitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	"github.com/yangwenmai/ratelimit/simpleratelimit"
)

type funcGetMonitorDatasByPrimaryKeys func([]string,
	string,
	int64,
	int64,
	int64) (map[string]map[int64]float64, error)

type funcGetMonitorDatasByMultiKeys func(map[string]interface{},
	string,
	int64,
	int64,
	int64) (map[int64]float64, error)

/*
	How many instances of information can be read per request
	Remote server limit, max is  10
*/
const monitorMaxRequestSize = 10

const timeNormalLayout = "2006-01-02 15:04:05"

const errorLogTemplate = "api[%s] fail, request body [%s], reason[%s]"

const debugLogTemplate = "api[%s] success, request body [%s], response body[%s]"

var rateLimit *simpleratelimit.RateLimiter

/*
	Each product has a function that gets metrics
	Format is: [productName]=>function
	eg:  funcGets["mysql"]=mysqlGetMonitorData
*/
var funcGetPrimaryKeys = map[string]funcGetMonitorDatasByPrimaryKeys{}

var funcGetMultiKeys = map[string]funcGetMonitorDatasByMultiKeys{}

var PrimaryKeys = map[string]string{}

/*
	api client for tentcent qcloud monitor.
*/
var remoteClient *remoteMonitor.Client

func splitSlice(slice []string, size int) [][]string {
	if size < 1 {
		size = 10
	}
	length := len(slice)
	chunks := int(math.Ceil(float64(length) / float64(size)))
	var buckets [][]string
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size
		if end > length {
			end = length
		}
		buckets = append(buckets, slice[i*size:end])
		i++
	}
	return buckets
}

/*
	get FuncMonitorData for the product, can get many instances info
*/
func GetPrimaryKeyFunc(productName string) funcGetMonitorDatasByPrimaryKeys {
	return funcGetPrimaryKeys[productName]
}

/*
	get FuncMonitorData for the product, only can get one specail dimensions info
*/
func GetMultiKeyFunc(productName string) funcGetMonitorDatasByMultiKeys {
	return funcGetMultiKeys[productName]
}

func InitClient(credentialConfig config.TencentCredential, rateLimitNumber int64) (errRet error) {
	credential := common.NewCredential(
		credentialConfig.AccessKey,
		credentialConfig.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 10

	client, err := remoteMonitor.NewClient(credential, credentialConfig.Region, cpf)
	if err != nil {
		errRet = err
		return
	}
	rateLimit = simpleratelimit.New(int(rateLimitNumber), time.Second)
	remoteClient = client
	return
}

func rateLimitCheck() {
	var sleepCount = 0
	for rateLimit.Limit() {
		sleepCount++
		if sleepCount > 10000 {
			log.Warnf("rate_limit sleep too much.")
			break
		}
		time.Sleep(time.Microsecond)
	}
	if sleepCount > 0 {
		log.Warnf("Hit rate_limit logic, a total of time delay is %f seconds", float64(sleepCount)/1000)
	}
}

/*
	get monitor data from qcloud, all  dimensions used to determine the key
	if func return 'errRet!=nil', allDataRet is empty, diff from getMonitorDataByPrimarykey
*/
func getMonitorDataByMultipleKeys(namespaceInMonitor string,
	dimensions map[string]interface{},
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64) (allDataRet map[int64]float64, errRet error) {

	if remoteClient == nil {
		errRet = fmt.Errorf("error,remote monitor client is not  initialized yet")
		log.Errorf(errRet.Error())
		return
	}

	rangeCount := rangeSeconds/periodSeconds - 1
	if rangeCount < 0 {
		rangeCount = 0
	}

	var (
		endTime             = ((time.Now().Unix() - delaySeconds) / periodSeconds) * periodSeconds
		startTime           = endTime - rangeCount*periodSeconds
		periodSecondsUint64 = uint64(periodSeconds)
	)

	startTimeStr := time.Unix(startTime, 0).Format(timeNormalLayout)
	endTimeStr := time.Unix(endTime, 0).Format(timeNormalLayout)

	request := remoteMonitor.NewGetMonitorDataRequest()
	request.Namespace = &namespaceInMonitor
	request.MetricName = &metricName
	request.Period = &periodSecondsUint64
	request.StartTime = &startTimeStr
	request.EndTime = &endTimeStr

	var requestDimensions = make([]*remoteMonitor.Dimension, 0, len(dimensions))
	var instance remoteMonitor.Instance

	for dName, dValue := range dimensions {
		if len(dName) == 0 {
			continue
		}
		var (
			strName   = dName
			strValue  = fmt.Sprintf("%v", dValue)
			dimension remoteMonitor.Dimension
		)
		dimension.Name = &strName
		dimension.Value = &strValue
		requestDimensions = append(requestDimensions, &dimension)
	}
	instance.Dimensions = requestDimensions
	request.Instances = []*remoteMonitor.Instance{&instance}

	rateLimitCheck()

	response, err := remoteClient.GetMonitorData(request)
	if err != nil {
		errRet = err
		log.Errorf(errorLogTemplate, request.GetAction(), request.ToJsonString(), errRet.Error())
		return
	} else {
		log.Debugf(debugLogTemplate, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}

	if len(response.Response.DataPoints) == 0 {
		log.Warnf("[Response.DataPoints].The product dimensions  has not monitor data at this time yet[instance is not use or not exist or not report to monitor now], monitor return: %s", response.ToJsonString())
		return
	}

	dataPoint := response.Response.DataPoints[0]

	if len(dataPoint.Values) == 0 {
		log.Warnf("[Response.DataPoints.Values].The product dimensions  has not monitor data at this time yet[instance is not use or not exist or not report to monitor now], monitor return: %s", response.ToJsonString())
		return
	}

	if len(dataPoint.Timestamps) != len(dataPoint.Values) {
		errRet = fmt.Errorf("api return error, len(Response.DataPoints.Dimensions.Timestamps) != len(Response.DataPoints.Dimensions.Values) ")
		log.Errorf(errorLogTemplate, request.GetAction(), request.ToJsonString(), errRet.Error())
		return
	}
	allDataRet = make(map[int64]float64, len(dataPoint.Timestamps))
	for index, value := range dataPoint.Timestamps {
		allDataRet[int64(*value)] = *(dataPoint.Values[index])
	}
	return
}

/*
get monitor data from qcloud, the primary key (eg:InstanceId) is the information used to determine the instance

if func return 'errRet!=nil', we also have to deal with it 'allDataRet'
*/
func getMonitorDataByPrimarykey(namespaceInMonitor string,
	primaryKeys []string,
	primaryKeyName string,
	metricName string,
	periodSeconds int64,
	rangeSeconds int64,
	delaySeconds int64) (allDataRet map[string]map[int64]float64, errRet error) {

	if remoteClient == nil {
		errRet = fmt.Errorf("error,remote monitor client is not  initialized yet")
		log.Errorf(errRet.Error())
		return
	}

	rangeCount := rangeSeconds/periodSeconds - 1
	if rangeCount < 0 {
		rangeCount = 0
	}

	allDataRet = make(map[string]map[int64]float64)

	var (
		endTime             = ((time.Now().Unix() - delaySeconds) / periodSeconds) * periodSeconds
		startTime           = endTime - rangeCount*periodSeconds
		periodSecondsUint64 = uint64(periodSeconds)
	)

	startTimeStr := time.Unix(startTime, 0).Format(timeNormalLayout)
	endTimeStr := time.Unix(endTime, 0).Format(timeNormalLayout)

	request := remoteMonitor.NewGetMonitorDataRequest()
	request.Namespace = &namespaceInMonitor
	request.MetricName = &metricName
	request.Period = &periodSecondsUint64
	request.StartTime = &startTimeStr
	request.EndTime = &endTimeStr

	buckets := splitSlice(primaryKeys, monitorMaxRequestSize)

	for _, bucket := range buckets {
		request.Instances = make([]*remoteMonitor.Instance, 0, len(bucket))
		for index := range bucket {
			var dimension remoteMonitor.Dimension
			var instance remoteMonitor.Instance
			dimension.Name = &primaryKeyName
			dimension.Value = &bucket[index]
			instance.Dimensions = []*remoteMonitor.Dimension{&dimension}
			request.Instances = append(request.Instances, &instance)
		}
		rateLimitCheck()
		response, err := remoteClient.GetMonitorData(request)

		if err != nil {
			errRet = err
			log.Errorf(errorLogTemplate, request.GetAction(), request.ToJsonString(), errRet.Error())
			continue
		} else {
			log.Debugf(debugLogTemplate, request.GetAction(), request.ToJsonString(), response.ToJsonString())
		}

		for _, dataPoint := range response.Response.DataPoints {
			if len(dataPoint.Dimensions) != 1 {
				errRet = fmt.Errorf("api return error, Response.DataPoints.Dimensions too long")
				log.Errorf(errorLogTemplate, request.GetAction(), request.ToJsonString(), errRet.Error())
				continue
			}
			if *(dataPoint.Dimensions[0].Name) != primaryKeyName {
				errRet = fmt.Errorf("api return error, Response.DataPoints.Dimensions not return %s", primaryKeyName)
				log.Errorf(errorLogTemplate, request.GetAction(), request.ToJsonString(), errRet.Error())
				continue
			}

			if len(dataPoint.Values) == 0 {
				log.Warnf("some product instances has not monitor data at this time yet[instance is not use or not exist or not report to monitor now], instance is: %s", *dataPoint.Dimensions[0].Value)
			}

			returnPrimaryKey := *(dataPoint.Dimensions[0].Value)
			if len(dataPoint.Timestamps) != len(dataPoint.Values) {
				errRet = fmt.Errorf("api return error, len(Response.DataPoints.Dimensions.Timestamps) != len(Response.DataPoints.Dimensions.Values) ")
				log.Errorf(errorLogTemplate, request.GetAction(), request.ToJsonString(), errRet.Error())
				continue
			}
			oneInstanceResult := make(map[int64]float64, len(dataPoint.Timestamps))
			for index, value := range dataPoint.Timestamps {
				oneInstanceResult[int64(*value)] = *(dataPoint.Values[index])
			}
			allDataRet[returnPrimaryKey] = oneInstanceResult
		}
	}
	return
}
