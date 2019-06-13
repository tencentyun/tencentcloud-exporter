package instances

import (
	"fmt"

	"github.com/tencentyun/tencentcloud-exporter/config"
)

type FuncGetInstanceIds func(map[string]interface{}) (map[string]map[string]interface{}, error)

/*
	Each product has a function that gets instance infos.
	Format is: [productName]=>function
	eg:  funcGets["mysql"]=getMysqlInstancesIds
*/

var funcGets = map[string]FuncGetInstanceIds{}

var credentialConfig config.TencentCredential

func GetInstanceFunc(productName string) FuncGetInstanceIds {
	return funcGets[productName]
}

func InitClient(cConfig config.TencentCredential) (errRet error) {
	credentialConfig = cConfig
	return
}

/*
   AbcdEfgh ---> abcd_efgh
*/
func ToUnderlineLower(s string) string {
	var interval byte = 'a' - 'A'
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += interval
			if i != 0 {
				b = append(b, '_')
			}
		}
		b = append(b, c)
	}
	return string(b)
}

/*
	Conversion of non-standard int64 api return to our data map.
	eg:api return UniqVpcId  we need format to  VpcId
*/
func setNonStandardInt64(needName string, int64Ptr *int64, toMap map[string]interface{}) {
	if int64Ptr == nil {
		return
	}
	toMap[needName] = int64(*int64Ptr)
}

/*
	Conversion of non-standard string api return to our data map.
*/
func setNonStandardStr(needName string, strPtr *string, toMap map[string]interface{}) {
	if strPtr == nil {
		return
	}
	toMap[needName] = *strPtr
}

/*
	check if the product instance satisfies the select rules defined by yml
*/
func meetConditions(dimensionSelect map[string]interface{}, productData map[string]interface{}, skipFilters map[string]bool) (ret bool) {
	for k, v := range dimensionSelect {
		if skipFilters[k] {
			continue
		}
		if productData[k] != nil {
			if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", productData[k]) {
				continue
			} else {
				return false
			}
		}
	}
	return true
}
