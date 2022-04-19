package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
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
func NewCMQTopicHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &cmqTopicHandler{
		baseProductHandler{
			monitorQueryKey: CMQTopicInstanceIDKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
