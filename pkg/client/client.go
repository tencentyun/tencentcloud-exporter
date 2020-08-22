package client

import (
	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	dc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
	mongodb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mongodb/v20190725"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
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
