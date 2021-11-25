package collector

import (
	"github.com/go-kit/kit/log"
)

const (
	CMQTopicNamespace     = "QCE/CMQTOPIC"
	CMQTopicInstanceIDKey = "topicId"
)

func init() {
	registerHandler(CMQTopicNamespace, defaultHandlerEnabled, NewCMQTopicHandler)
}

type cmqTopicHandler struct {
	baseProductHandler
}

func (h *cmqTopicHandler) GetNamespace() string {
	return CMQTopicNamespace
}
func NewCMQTopicHandler(c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &cmqTopicHandler{
		baseProductHandler{
			monitorQueryKey: CMQTopicInstanceIDKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
