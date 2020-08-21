package metric

import (
	"fmt"
	"github.com/tencentyun/tencentcloud-exporter/pkg/instance"
)

type TcmSeries struct {
	Id          string
	Metric      *TcmMetric
	QueryLabels Labels
	Instance    instance.TcInstance
}

func GetTcmSeriesId(m *TcmMetric, ql Labels) (string, error) {
	m5, e := ql.Md5()
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%s-%s", m.Id, m5), nil
}

func NewTcmSeries(m *TcmMetric, ql Labels, ins instance.TcInstance) (*TcmSeries, error) {
	id, err := GetTcmSeriesId(m, ql)
	if err != nil {
		return nil, err
	}

	s := &TcmSeries{
		Id:          id,
		Metric:      m,
		QueryLabels: ql,
		Instance:    ins,
	}
	return s, nil

}
