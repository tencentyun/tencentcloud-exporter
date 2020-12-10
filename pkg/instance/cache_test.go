package instance

import (
	"github.com/prometheus/common/promlog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
)

func TestTcRedisInstanceNodeCache_GetNodeInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNodeRepo := NewMockRedisTcInstanceNodeRepository(ctrl)

	loglevel := &promlog.AllowedLevel{}
	loglevel.Set("debug")
	logformat := &promlog.AllowedFormat{}
	logformat.Set("logfmt")

	cache := &TcRedisInstanceNodeCache{
		Raw:            mockNodeRepo,
		cache:          map[string]*sdk.DescribeInstanceNodeInfoResponse{},
		lastReloadTime: map[string]time.Time{},
		reloadInterval: 3 * time.Minute,
		logger: promlog.New(&promlog.Config{
			Level:  loglevel,
			Format: logformat,
		}),
	}

	mockInstanceId := "crs-12345678"

	// case 1: get from api, init
	mockNodeRepo.EXPECT().GetNodeInfo(gomock.Any()).Return(&sdk.DescribeInstanceNodeInfoResponse{}, nil)
	_, err := cache.GetNodeInfo(mockInstanceId)
	assert.NoError(t, err)

	// case 2: get from api, expire
	mockCacheNode := &sdk.DescribeInstanceNodeInfoResponse{}
	cache.cache[mockInstanceId] = mockCacheNode
	cache.lastReloadTime[mockInstanceId] = time.Now().Add(-4 * time.Minute)
	mockNodeRepo.EXPECT().GetNodeInfo(gomock.Any()).Return(&sdk.DescribeInstanceNodeInfoResponse{}, nil)
	node, err := cache.GetNodeInfo(mockInstanceId)
	assert.NoError(t, err)

	// case 3: get from cache
	cache.cache[mockInstanceId] = &sdk.DescribeInstanceNodeInfoResponse{}
	cache.lastReloadTime[mockInstanceId] = time.Now().Add(-1 * time.Minute)

	node, err = cache.GetNodeInfo(mockInstanceId)
	assert.NoError(t, err)
	assert.Equal(t, node, cache.cache[mockInstanceId])
}
