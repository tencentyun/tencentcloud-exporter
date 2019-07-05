package instances

import (
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/tencentyun/tencentcloud-exporter/config"
	"os"
	"strings"
	"testing"
)

func TestGetNatInstancesIds(t *testing.T) {

	credentialConfig.AccessKey = os.Getenv(config.EnvAccessKey)
	credentialConfig.SecretKey = os.Getenv(config.EnvSecretKey)
	credentialConfig.Region = os.Getenv(config.EnvRegion)

	if credentialConfig.AccessKey == "" || credentialConfig.SecretKey == "" || credentialConfig.Region == "" {
		log.Errorf("should set env TENCENTCLOUD_SECRET_ID , TENCENTCLOUD_SECRET_KEY and TENCENTCLOUD_REGION")
		return
	}

	fun := GetInstanceFunc(NatProductName)
	if fun == nil {
		t.Errorf("%s get InstancesIds func not exist.", NatProductName)
		return
	}
	_, err := fun(nil)
	if err != nil {
		t.Errorf("%s get InstancesIds func fail,reason %s", NatProductName, err.Error())
		return
	}

	filters := make(map[string]interface{})

	filters["InstanceName"] = "t"

	instances, err := fun(filters)
	if err != nil {
		t.Errorf("%s get InstancesIds func fail,reason %s", NatProductName, err.Error())
		return
	}
	for instanceId,instance:=range instances{
		if !strings.Contains(fmt.Sprintf("%+v",instance["InstanceName"]),
			fmt.Sprintf("%+v",filters["InstanceName"])){

			t.Errorf("%s get InstancesIds return[%s] not match filters", NatProductName, instanceId)
			return
		}
	}
}