package client

import (
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	kafka "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ckafka/v20190819"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	cmq "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cmq/v20190304"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	dc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
	es "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/es/v20180416"
	lh "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	mariadb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mariadb/v20170312"
	memcached "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/memcached/v20190318"
	mongodb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mongodb/v20190725"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	pg "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/postgres/v20170312"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	sqlserver "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sqlserver/v20180328"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
)

func NewMonitorClient(conf *config.TencentConfig) (*monitor.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "monitor.tencentcloudapi.com"
	return monitor.NewClient(credential, conf.Credential.Region, cpf)
}

func NewMongodbClient(conf *config.TencentConfig) (*mongodb.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "mongodb.tencentcloudapi.com"
	return mongodb.NewClient(credential, conf.Credential.Region, cpf)
}

func NewCdbClient(conf *config.TencentConfig) (*cdb.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cdb.tencentcloudapi.com"
	return cdb.NewClient(credential, conf.Credential.Region, cpf)
}

func NewCvmClient(conf *config.TencentConfig) (*cvm.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	return cvm.NewClient(credential, conf.Credential.Region, cpf)
}

func NewRedisClient(conf *config.TencentConfig) (*redis.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "redis.tencentcloudapi.com"
	return redis.NewClient(credential, conf.Credential.Region, cpf)
}

func NewDcClient(conf *config.TencentConfig) (*dc.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dc.tencentcloudapi.com"
	return dc.NewClient(credential, conf.Credential.Region, cpf)
}

func NewClbClient(conf *config.TencentConfig) (*clb.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "clb.tencentcloudapi.com"
	return clb.NewClient(credential, conf.Credential.Region, cpf)
}

func NewVpvClient(conf *config.TencentConfig) (*vpc.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	return vpc.NewClient(credential, conf.Credential.Region, cpf)
}

func NewCbsClient(conf *config.TencentConfig) (*cbs.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cbs.tencentcloudapi.com"
	return cbs.NewClient(credential, conf.Credential.Region, cpf)
}

func NewSqlServerClient(conf *config.TencentConfig) (*sqlserver.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sqlserver.tencentcloudapi.com"
	return sqlserver.NewClient(credential, conf.Credential.Region, cpf)
}

func NewMariaDBClient(conf *config.TencentConfig) (*mariadb.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	return mariadb.NewClient(credential, conf.Credential.Region, cpf)
}

func NewESClient(conf *config.TencentConfig) (*es.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	return es.NewClient(credential, conf.Credential.Region, cpf)
}

func NewCMQClient(conf *config.TencentConfig) (*cmq.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	return cmq.NewClient(credential, conf.Credential.Region, cpf)
}

func NewPGClient(conf *config.TencentConfig) (*pg.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	return pg.NewClient(credential, conf.Credential.Region, cpf)
}

func NewMemcacheClient(conf *config.TencentConfig) (*memcached.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	return memcached.NewClient(credential, conf.Credential.Region, cpf)
}

func NewLighthouseClient(conf *config.TencentConfig) (*lh.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	return lh.NewClient(credential, conf.Credential.Region, cpf)
}

func NewKafkaClient(conf *config.TencentConfig) (*kafka.Client, error) {
	credential := common.NewCredential(
		conf.Credential.AccessKey,
		conf.Credential.SecretKey,
	)
	cpf := profile.NewClientProfile()
	return kafka.NewClient(credential, conf.Credential.Region, cpf)
}
