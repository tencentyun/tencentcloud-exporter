package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const DefaultPeriodSeconds = 60

const DefaultDelaySeconds = 300

type TencentCredential struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
}
type TencentMetric struct {
	Namespace     string                 `yaml:"tc_namespace"`
	MetricName    string                 `yaml:"tc_metric_name"`
	MetricReName  string                 `yaml:"tc_metric_rename"`
	Labels        []string               `yaml:"tc_labels"`
	Dimensions    map[string]interface{} `yaml:"tc_myself_dimensions"`
	Filters       map[string]interface{} `yaml:"tc_filters"`
	Statistics    []string               `yaml:"tc_statistics"`
	PeriodSeconds int64                  `yaml:"period_seconds"`
	RangeSeconds  int64                  `yaml:"range_seconds"`
	DelaySeconds  int64                  `yaml:"delay_seconds"`
}

type TencentConfig struct {
	Credential TencentCredential `yaml:"credential"`
	Metrics    []TencentMetric   `yaml:"metrics"`
	RateLimit  int64             `yaml:"rate_limit"`
	filename   string
}

func NewConfig() *TencentConfig {
	return &TencentConfig{}
}

func (me *TencentConfig) LoadFile(filename string) (errRet error) {
	me.filename = filename
	content, err := ioutil.ReadFile(me.filename)
	if err != nil {
		errRet = err
		return
	}
	if err = yaml.UnmarshalStrict(content, me); err != nil {
		errRet = err
		return
	}
	if errRet = me.check(); errRet != nil {
		return errRet
	}
	me.fillDefault()
	return nil
}

func (me *TencentConfig) check() (errRet error) {
	if me.Credential.AccessKey == "" ||
		me.Credential.SecretKey == "" ||
		me.Credential.Region == "" {
		return fmt.Errorf("error, missing credential information!")
	}
	for _, metric := range me.Metrics {
		if metric.MetricName == "" {
			return fmt.Errorf("error, missing tc_metric_name !")
		}
		if len(strings.Split(metric.Namespace, `/`)) != 2 {
			return fmt.Errorf("error, tc_namespace should be 'xxxxxx/productName' format")
		}
		if len(metric.Dimensions) != 0 && (len(metric.Filters) != 0 || len(metric.Labels) != 0) {
			return fmt.Errorf("error, [tc_myself_dimensions] conflict with [tc_labels,tc_filters]")
		}
	}

	return nil
}

func (me *TencentConfig) fillDefault() {

	if me.RateLimit <= 0 {
		me.RateLimit = 10
	}

	for index, metric := range me.Metrics {
		if metric.PeriodSeconds == 0 {
			me.Metrics[index].PeriodSeconds = DefaultPeriodSeconds
		}
		if metric.DelaySeconds == 0 {
			me.Metrics[index].DelaySeconds = DefaultDelaySeconds
		}

		if metric.RangeSeconds == 0 {
			metric.RangeSeconds = metric.PeriodSeconds
		}

		if metric.MetricReName == "" {
			me.Metrics[index].MetricReName = me.Metrics[index].MetricName
		}
	}
}
