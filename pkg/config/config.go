package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tencentyun/tencentcloud-exporter/pkg/constant"
	"github.com/tencentyun/tencentcloud-exporter/pkg/util"
	"gopkg.in/yaml.v2"
)

const (
	DefaultPeriodSeconds        = 60
	DefaultDelaySeconds         = 300
	DefaultRelodIntervalMinutes = 60
	DefaultRateLimit            = 15
	DefaultQueryMetricBatchSize = 50

	EnvAccessKey = "TENCENTCLOUD_SECRET_ID"
	EnvSecretKey = "TENCENTCLOUD_SECRET_KEY"
	EnvRegion    = "TENCENTCLOUD_REGION"
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
	}

	SupportStatisticsTypes = map[string]bool{
		"max":  true,
		"min":  true,
		"avg":  true,
		"last": true,
	}
)

type TencentCredential struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
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
	RelodIntervalMinutes  int64               `yaml:"relod_interval_minutes"`
}

func (p *TencentProduct) IsReloadEnable() bool {
	if len(p.OnlyIncludeMetrics) > 0 {
		return false
	}
	if util.IsStrInList(constant.NotSupportInstanceNamespaces, p.Namespace) {
		return false
	}
	return p.AllInstances
}

type TencentConfig struct {
	Credential           TencentCredential `yaml:"credential"`
	Metrics              []TencentMetric   `yaml:"metrics"`
	Products             []TencentProduct  `yaml:"products"`
	RateLimit            float64           `yaml:"rate_limit"`
	MetricQueryBatchSize int               `yaml:"metric_query_batch_size"`
	Filename             string            `yaml:"filename"`
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
	if err = yaml.UnmarshalStrict(content, c); err != nil {
		return err

	}
	if err = c.check(); err != nil {
		return err
	}
	c.fillDefault()
	return nil
}

func (c *TencentConfig) check() (err error) {
	if c.Credential.AccessKey == "" {
		c.Credential.AccessKey = os.Getenv(EnvAccessKey)
		if c.Credential.AccessKey == "" {
			return fmt.Errorf("credential.access_key is empty, must be set")
		}
	}
	if c.Credential.SecretKey == "" {
		c.Credential.SecretKey = os.Getenv(EnvSecretKey)
		if c.Credential.SecretKey == "" {
			return fmt.Errorf("credential.secret_key is empty, must be set")
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
		if product.RelodIntervalMinutes <= 0 {
			c.Products[index].RelodIntervalMinutes = DefaultRelodIntervalMinutes
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
