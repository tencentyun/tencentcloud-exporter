package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/common"
)

const (
	MariaDBNamespace     = "QCE/MARIADB"
	MariaDBInstanceIDKey = "InstanceId"
)

func init() {
	registerHandler(MariaDBNamespace, defaultHandlerEnabled, NewMariaDBHandler)
}

type mariaDBHandler struct {
	baseProductHandler
}

func (h *mariaDBHandler) GetNamespace() string {
	return MariaDBNamespace
}
func NewMariaDBHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &mariaDBHandler{
		baseProductHandler{
			monitorQueryKey: MariaDBInstanceIDKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
