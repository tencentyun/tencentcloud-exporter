package instance

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

var (
	factoryMap = make(map[string]func(*config.TencentConfig, log.Logger) (TcInstanceRepository, error))
)

type TcInstanceRepository interface {
	GetInstanceKey() string
	Get(id string) (TcInstance, error)
	ListByIds(ids []string) ([]TcInstance, error)
	ListByFilters(filters map[string]string) ([]TcInstance, error)
}

func NewTcInstanceRepository(namespace string, conf *config.TencentConfig, logger log.Logger) (TcInstanceRepository, error) {
	f, exists := factoryMap[namespace]
	if !exists {
		return nil, fmt.Errorf("Namespace not support, namespace=%s ", namespace)
	}
	return f(conf, logger)
}

func registerRepository(namespace string, factory func(*config.TencentConfig, log.Logger) (TcInstanceRepository, error)) {
	factoryMap[namespace] = factory
}
