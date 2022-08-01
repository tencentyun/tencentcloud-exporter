package collector

import (
	"github.com/go-kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
)

const (
	ESNamespace     = "QCE/CES"
	ESInstanceIDKey = "uInstanceId"
)

func init() {
	registerHandler(ESNamespace, defaultHandlerEnabled, NewESHandler)
}

type esHandler struct {
	baseProductHandler
}

func (h *esHandler) GetNamespace() string {
	return ESNamespace
}

func NewESHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &esHandler{
		baseProductHandler{
			monitorQueryKey: ESInstanceIDKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
