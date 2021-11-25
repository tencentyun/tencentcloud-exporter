package collector

import (
	"github.com/go-kit/kit/log"
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

func NewESHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &esHandler{
		baseProductHandler{
			monitorQueryKey: ESInstanceIDKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
