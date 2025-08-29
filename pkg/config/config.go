package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tencentyun/tencentcloud-exporter/pkg/constant"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
	"gopkg.in/yaml.v2"
	"github.com/kobyt2/common-services/utils/aes"
)

const (
	DefaultPeriodSeconds         = 60
	DefaultDelaySeconds          = 300
	DefaultReloadIntervalMinutes = 60
	DefaultRateLimit             = 15
	DefaultQueryMetricBatchSize  = 50

	EnvAccessKey   = "TENCENTCLOUD_SECRET_ID"
	EnvSecretKey   = "TENCENTCLOUD_SECRET_KEY"
	EnvServiceRole = "TENCENTCLOUD_SERVICE_ROLE"
	EnvRegion      = "TENCENTCLOUD_REGION"
)

var (
	Product2Namespace = map[string]string{
		"cmongo":        "QCE/CMONGO",
		"mongo":         "QCE/CMONGO",
		"cdb":           "QCE/CDB",
		"mysql":         "QCE/CDB",
		"cvm":           "QCE/CVM",
		"redis":         "QCE/REDIS",
		"redis_cluster": "QCE/REDIS",
		"redis_mem":     "QCE/REDIS_MEM",
		"cluster_redis": "QCE/REDIS",
		"dc":            "QCE/DC",
		"dcx":           "QCE/DCX",
		"lb_public":     "QCE/LB_PUBLIC",
		"public_clb":    "QCE/LB_PUBLIC",
		"loadbalance":   "QCE/LOADBALANCE",
		"7layer_clb":    "QCE/LOADBALANCE",
		"nat_gateway":   "QCE/NAT_GATEWAY",
		"nat":           "QCE/NAT_GATEWAY",
		"cos":           "QCE/COS",
		"cdn":           "QCE/CDN",
		"cbs":           "QCE/BLOCK_STORAGE",
		"sqlserver":     "QCE/SQLSERVER",
		"mariadb":       "QCE/MARIADB",
		"ces":           "QCE/CES",
		"cmq":           "QCE/CMQ",
		"cmqtopic":      "QCE/CMQTOPIC",
		"postgres":      "QCE/POSTGRES",
		"memcached":     "QCE/MEMCACHED",
		"lighthouse":    "QCE/LIGHTHOUSE",
		"ckafka":        "QCE/CKAFKA",
		"tdmysql":       "QCE/TDMYSQL",
		"lb":            "QCE/LB",            // for eip
		"tdmq":          "QCE/TDMQ",          // for tdmq
		"zookeeper":     "TSE/ZOOKEEPER",     // for zookeeper
		"nacos":         "TSE/NACOS",         // for nacos
		"vpngw":         "QCE/VPNGW",         // for vpngw
		"vpnx":          "QCE/VPNX",          // for vpnx
		"cynosdb_mysql": "QCE/CYNOSDB_MYSQL", // for cynosdb_mysql
		"dts":           "QCE/DTS",           // for dts
		"dcg":           "QCE/DCG",           // for dcg
		"vbc":           "QCE/VBC",           // for vbc
		"qaap":          "QCE/QAAP",          // for qaap
		"lb_private":    "QCE/LB_PRIVATE",
		"waf":           "QCE/WAF",
		"cfs":           "QCE/CFS",
	}

	SupportStatisticsTypes = map[string]bool{
		"max":  true,
		"min":  true,
		"avg":  true,
		"last": true,
	}
)

var QcloudNamespace = []string{"COS", "CDN", "QAAP", "WAF"}

type TencentCredential struct {
	AccessKey   string `yaml:"access_key"`
	SecretKey   string `yaml:"secret_key"`
	EncryptionKey string `yaml:"encryption_key"`
	Role        string `yaml:"role"`
	Region      string `yaml:"region"`
	Token       string `yaml:"token"`
	ExpiredTime int64  `yaml:"expired_time"`
	IsInternal  bool   `yaml:"is_internal"`
}

type TencentMetric struct {
	Namespace      string            `yaml:"tc_namespace"`
	MetricName     string            `yaml:"tc_metric_name"`
	MetricReName   string            `yaml:"tc_metric_rename"`
	MetricNameType int32             `yaml:"tc_metric_name_type"` // 1=大写转下划线, 2=全小写
	Labels         []string          `yaml:"tc_labels"`
	Dimensions     map[string]string `yaml:"tc_myself_dimensions"`
	Filters        map[string]string `yaml:"tc_filters"`
	Statistics     []string          `yaml:"tc_statistics"`
	PeriodSeconds  int64             `yaml:"period_seconds"`
	RangeSeconds   int64             `yaml:"range_seconds"`
	DelaySeconds   int64             `yaml:"delay_seconds"`
}

type TencentProduct struct {
	Namespace             string              `yaml:"namespace"`
	AllMetrics            bool                `yaml:"all_metrics"`
	AllInstances          bool                `yaml:"all_instances"`
	ExtraLabels           []string            `yaml:"extra_labels"`
	OnlyIncludeMetrics    []string            `yaml:"only_include_metrics"`
	ExcludeMetrics        []string            `yaml:"exclude_metrics"`
	InstanceFilters       map[string]string   `yaml:"instance_filters"`
	OnlyIncludeInstances  []string            `yaml:"only_include_instances"`
	ExcludeInstances      []string            `yaml:"exclude_instances"`
	CustomQueryDimensions []map[string]string `yaml:"custom_query_dimensions"`
	Statistics            []string            `yaml:"statistics_types"`
	PeriodSeconds         int64               `yaml:"period_seconds"`
	RangeSeconds          int64               `yaml:"range_seconds"`
	DelaySeconds          int64               `yaml:"delay_seconds"`
	MetricNameType        int32               `yaml:"metric_name_type"` // 1=大写转下划线, 2=全小写
	ReloadIntervalMinutes int64               `yaml:"reload_interval_minutes"`
}

type metadataResponse struct {
	TmpSecretId  string
	TmpSecretKey string
	Token        string
	ExpiredTime  int64
	Code         string
}

func (p *TencentProduct) IsReloadEnable() bool {
	if util.IsStrInList(constant.NotSupportInstanceNamespaces, p.Namespace) {
		return false
	}
	return true
}

type TencentConfig struct {
	Credential           TencentCredential `yaml:"credential"`
	Metrics              []TencentMetric   `yaml:"metrics"`
	Products             []TencentProduct  `yaml:"products"`
	RateLimit            float64           `yaml:"rate_limit"`
	MetricQueryBatchSize int               `yaml:"metric_query_batch_size"`
	Filename             string            `yaml:"filename"`
	CacheInterval        int64             `yaml:"cache_interval"`   // 单位 s
	IsInternational      bool              `yaml:"is_international"` // true 表示是国际站
}

func NewConfig() *TencentConfig {
	return &TencentConfig{}
}

func (c *TencentConfig) LoadFile(filename string) error {
    c.Filename = filename
    content, err := ioutil.ReadFile(c.Filename)
    if err != nil {
        return err
    }

    // 解析 YAML 文件
    if err = yaml.UnmarshalStrict(content, c); err != nil {
        return err
    }

    // 假设加密密钥存储在 YAML 中
    crypto := aes.NewCryptoDB(c.Credential.EncryptionKey) // 直接读取配置中的密钥

    // 解密 access_key 和 secret_key
    decryptedAccessKey := crypto.Decrypt(c.Credential.AccessKey)
    decryptedSecretKey := crypto.Decrypt(c.Credential.SecretKey)

    c.Credential.AccessKey = decryptedAccessKey
    c.Credential.SecretKey = decryptedSecretKey

    // 检查配置文件是否合法
    if err = c.check(); err != nil {
        return err
    }

    c.fillDefault()
    return nil
}


func (c *TencentConfig) check() (err error) {
	if c.Credential.AccessKey == "" {
		c.Credential.AccessKey = os.Getenv(EnvAccessKey)
	}
	if c.Credential.SecretKey == "" {
		c.Credential.SecretKey = os.Getenv(EnvSecretKey)
	}
	if c.Credential.AccessKey != "" && c.Credential.SecretKey != "" {
		c.Credential.Role = "" // 优先使用密钥，根据 role 是否为空判断使用密钥还是 role
	} else if c.Credential.Role == "" {
		c.Credential.Role = os.Getenv(EnvServiceRole)
		if c.Credential.Role == "" {
			return fmt.Errorf("credential.access_key or credential.secret_key or credential.role is empty, must be set")
		}
	}

	if c.Credential.Region == "" {
		c.Credential.Region = os.Getenv(EnvRegion)
		if c.Credential.Region == "" {
			return fmt.Errorf("credential.region is empty, must be set")
		}
	}

	for _, mconf := range c.Metrics {
		if mconf.MetricName == "" {
			return fmt.Errorf("tc_metric_name is empty, must be set")
		}
		nsitems := strings.Split(mconf.Namespace, `/`)
		if len(nsitems) != 2 {
			return fmt.Errorf("tc_namespace should be 'xxxxxx/productName' format")
		}
		pname := nsitems[1]
		if _, exists := Product2Namespace[strings.ToLower(pname)]; !exists {
			return fmt.Errorf("tc_namespace productName not support")
		}
		for _, statistic := range mconf.Statistics {
			_, exists := SupportStatisticsTypes[strings.ToLower(statistic)]
			if !exists {
				return fmt.Errorf("statistic type not support, type=%s", statistic)
			}
		}
	}

	for _, pconf := range c.Products {
		nsitems := strings.Split(pconf.Namespace, `/`)
		if len(nsitems) != 2 {
			return fmt.Errorf("namespace should be 'xxxxxx/productName' format")
		}
		pname := nsitems[1]
		if _, exists := Product2Namespace[strings.ToLower(pname)]; !exists {
			return fmt.Errorf("namespace productName not support, %s", pname)
		}
		if len(pconf.OnlyIncludeInstances) == 0 && !pconf.AllInstances && len(pconf.CustomQueryDimensions) == 0 {
			return fmt.Errorf("must set all_instances or only_include_instances or custom_query_dimensions")
		}
	}

	return nil
}

func (c *TencentConfig) fillDefault() {
	if c.RateLimit <= 0 {
		c.RateLimit = DefaultRateLimit
	}

	if c.MetricQueryBatchSize <= 0 || c.MetricQueryBatchSize > 100 {
		c.MetricQueryBatchSize = DefaultQueryMetricBatchSize
	}

	for index, metric := range c.Metrics {
		if metric.PeriodSeconds == 0 {
			c.Metrics[index].PeriodSeconds = DefaultPeriodSeconds
		}
		if metric.DelaySeconds == 0 {
			c.Metrics[index].DelaySeconds = c.Metrics[index].PeriodSeconds
		}

		if metric.RangeSeconds == 0 {
			metric.RangeSeconds = metric.PeriodSeconds
		}

		if metric.MetricReName == "" {
			c.Metrics[index].MetricReName = c.Metrics[index].MetricName
		}
	}

	for index, product := range c.Products {
		if product.ReloadIntervalMinutes <= 0 {
			c.Products[index].ReloadIntervalMinutes = DefaultReloadIntervalMinutes
		}
	}
}

func (c *TencentConfig) GetNamespaces() (nps []string) {
	nsSet := map[string]struct{}{}
	for _, pconf := range c.Products {
		ns := GetStandardNamespaceFromCustomNamespace(pconf.Namespace)
		nsSet[ns] = struct{}{}
	}
	for _, mconf := range c.Metrics {
		ns := GetStandardNamespaceFromCustomNamespace(mconf.Namespace)
		nsSet[ns] = struct{}{}
	}

	for np := range nsSet {
		nps = append(nps, np)
	}
	return
}

func (c *TencentConfig) GetMetricConfigs(namespace string) (mconfigs []TencentMetric) {
	for _, mconf := range c.Metrics {
		ns := GetStandardNamespaceFromCustomNamespace(mconf.Namespace)
		if ns == namespace {
			mconfigs = append(mconfigs, mconf)
		}
	}
	return
}

func (c *TencentConfig) GetProductConfig(namespace string) (TencentProduct, error) {
	for _, pconf := range c.Products {
		ns := GetStandardNamespaceFromCustomNamespace(pconf.Namespace)
		if ns == namespace {
			return pconf, nil
		}
	}
	return TencentProduct{}, fmt.Errorf("namespace config not found")
}

func GetStandardNamespaceFromCustomNamespace(cns string) string {
	items := strings.Split(cns, "/")
	if len(items) != 2 {
		panic(fmt.Sprintf("Namespace should be 'customPrefix/productName' format"))
	}
	pname := items[1]
	sns, exists := Product2Namespace[strings.ToLower(pname)]
	if exists {
		return sns
	} else {
		panic(fmt.Sprintf("Product not support, namespace=%s, product=%s", cns, pname))
	}
}
