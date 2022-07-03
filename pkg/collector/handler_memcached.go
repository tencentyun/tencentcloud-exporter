package collector

import (
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
)

const (
	MemcachedNamespace     = "QCE/MEMCACHED"
	MemcachedInstanceIDKey = "instanceid"
)

func init() {
	registerHandler(MemcachedNamespace, defaultHandlerEnabled, NewMemcachedHandler)
}

type memcachedHandler struct {
	baseProductHandler
}

func (h *memcachedHandler) GetNamespace() string {
	return MemcachedNamespace
}

func NewMemcachedHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &memcachedHandler{
		baseProductHandler{
			monitorQueryKey: MemcachedInstanceIDKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
