package instance

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

var (
	factoryMap = make(map[string]func(*config.TencentConfig, log.Logger) (TcInstanceRepository, error))
)

// 每个产品的实例对象的Repository
type TcInstanceRepository interface {
	// 获取实例id
	GetInstanceKey() string
	// 根据id, 获取实例对象
	Get(id string) (TcInstance, error)
	// 根据id列表, 获取所有的实例对象
	ListByIds(ids []string) ([]TcInstance, error)
	// 根据filters, 获取符合条件的所有实例对象
	ListByFilters(filters map[string]string) ([]TcInstance, error)
}

func NewTcInstanceRepository(namespace string, conf *config.TencentConfig, logger log.Logger) (TcInstanceRepository, error) {
	f, exists := factoryMap[namespace]
	if !exists {
		return nil, fmt.Errorf("Namespace not support, namespace=%s ", namespace)
	}
	return f(conf, logger)
}

// 将TcInstanceRepository注册到factoryMap中
func registerRepository(namespace string, factory func(*config.TencentConfig, log.Logger) (TcInstanceRepository, error)) {
	factoryMap[namespace] = factory
}
