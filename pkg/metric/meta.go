package metric

import (
	"errors"
	"fmt"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
	"sort"
	"strconv"
	"strings"
)

type TcmMeta struct {
	Id                string
	Namespace         string
	ProductName       string
	MetricName        string
	SupportDimensions []string
	m                 *monitor.MetricSet
}

func (meta *TcmMeta) GetPeriod(confPeriod int64) (int64, error) {
	if len(meta.m.Period) == 0 {
		return 0, errors.New("period is empty")
	}

	var allowPeriods []int
	for _, p := range meta.m.Period {
		allowPeriods = append(allowPeriods, int(*p))
	}
	sort.Ints(allowPeriods)

	var period int64
	if confPeriod != 0 {
		period = confPeriod
	} else {
		period = config.DefaultPeriodSeconds
	}

	idx := sort.SearchInts(allowPeriods, int(period))
	if idx < len(allowPeriods) {
		return period, nil
	} else {
		return int64(allowPeriods[0]), nil
	}

}

func (meta *TcmMeta) GetStatType(period int64) (string, error) {
	var statType string
	var defaultStatType string
	for _, p := range meta.m.Periods {
		i, err := strconv.ParseInt(*p.Period, 10, 64)
		if err != nil {
			return "", err
		}
		if i == period {
			statType = *p.StatType[0]
		}
		if i == 300 {
			defaultStatType = *p.StatType[0]
		}
	}
	if statType != "" {
		return statType, nil
	}
	if defaultStatType != "" {
		return defaultStatType, nil
	}

	return "", fmt.Errorf("not found statType, period=%d", period)

}

func NewTcmMeta(m *monitor.MetricSet) (*TcmMeta, error) {
	id := fmt.Sprintf("%s-%s", *m.Namespace, *m.MetricName)

	var productName string
	nsItems := strings.Split(*m.Namespace, "/")
	productName = nsItems[len(nsItems)-1]

	var supportDimensions []string
	for _, dimension := range m.Dimensions {
		for _, d := range dimension.Dimensions {
			supportDimensions = append(supportDimensions, *d)
		}
	}

	meta := &TcmMeta{
		Id:                id,
		Namespace:         *m.Namespace,
		ProductName:       productName,
		MetricName:        *m.MetricName,
		SupportDimensions: supportDimensions,
		m:                 m,
	}
	return meta, nil

}
