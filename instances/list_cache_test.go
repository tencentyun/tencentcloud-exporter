package instances

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {

	oldMaxCacheTime := maxCacheTime

	defer func() {
		maxCacheTime = oldMaxCacheTime
	}()

	maxCacheTime = 3

	productName := "testProduct"

	nowUnixNano := time.Now().UnixNano()

	dimensionSelect := map[string]interface{}{}
	dimensionSelect["name"] = "guagua"
	dimensionSelect["age"] = 20

	key := getCacheKey(productName, dimensionSelect)

	value := map[string]map[string]interface{}{}

	value["info"] = map[string]interface{}{}
	value["info"]["id"] = nowUnixNano

	setCache(key, value)
	ret := getCache(key, true)

	checkOk := false

	if ret["info"] != nil && ret["info"]["id"] != nil {
		if id, ok := ret["info"]["id"].(int64); ok && id == nowUnixNano {
			checkOk = true
		}
	}
	if !checkOk {
		t.Errorf("getCacheKey return error,the data we save is missing or incorrect")
	}

	time.Sleep(time.Duration(maxCacheTime+1) * time.Second)

	if ret = getCache(key, true); ret != nil {
		t.Errorf("getCacheKey return error, cache is not expired")
	}

	ret = getCache(key, false)
	checkOk = false
	if ret["info"] != nil && ret["info"]["id"] != nil {
		if id, ok := ret["info"]["id"].(int64); ok && id == nowUnixNano {
			checkOk = true
		}
	}
	if !checkOk {
		t.Errorf("getCacheKey return error,in the non-strict case, our data also fiction, but should exist")
	}

	if ret = getCache(key+"a", false); ret != nil {
		t.Errorf("getCacheKey return error, the wrong key returns data")
	}

}
