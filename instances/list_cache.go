package instances

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type cacheItem struct {
	cacheData      map[string]map[string]interface{}
	lastModifyTime int64
}

/*Cache the list of products for equal conditions, cache && locker && time*/
var memCache = map[string]cacheItem{}
var memCacheLocker sync.Mutex
var maxCacheTime int64 = 40

func getCache(key string, strict bool) (ret map[string]map[string]interface{}) {

	memCacheLocker.Lock()
	defer memCacheLocker.Unlock()

	saveRet, ok := memCache[key]
	if !ok {
		return
	}

	if !strict || (saveRet.lastModifyTime+maxCacheTime > time.Now().Unix()) {
		ret = saveRet.cacheData
		return
	}

	return
}

func setCache(key string, value map[string]map[string]interface{}) {
	memCacheLocker.Lock()
	memCache[key] = cacheItem{
		cacheData:      value,
		lastModifyTime: time.Now().Unix(),
	}
	memCacheLocker.Unlock()
}

func getCacheKey(productName string, dimensionSelect map[string]interface{}) (key string) {
	var buf bytes.Buffer
	js, _ := json.Marshal(dimensionSelect)
	buf.Write(js)
	buf.WriteString(credentialConfig.AccessKey)
	buf.WriteString(credentialConfig.SecretKey)
	buf.WriteString(credentialConfig.Region)
	md := md5.New()
	md.Write(buf.Bytes())
	key = fmt.Sprintf("%s_%x", md.Sum(nil), productName)
	return
}
