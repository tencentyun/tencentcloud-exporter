package metric

import (
	"fmt"

	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

// 代表一个数据点
type TcmSample struct {
	Timestamp  float64
	Value      float64
	Dimensions []*monitor.Dimension
}

// 代表一个时间线的多个数据点
type TcmSamples struct {
	Series  *TcmSeries
	Samples []*TcmSample
}

func (s *TcmSamples) GetLatestPoint() (point *TcmSample, err error) {
	if len(s.Samples) == 1 {
		return s.Samples[0], nil
	} else {
		return s.Samples[len(s.Samples)-1], nil
	}
}

func (s *TcmSamples) GetMaxPoint() (point *TcmSample, err error) {
	maxValue := s.Samples[0].Value
	var maxIdx int
	for idx, sample := range s.Samples {
		if sample.Value > maxValue {
			maxValue = sample.Value
			maxIdx = idx
		}
	}
	return s.Samples[maxIdx], nil
}

func (s *TcmSamples) GetMinPoint() (point *TcmSample, err error) {
	minValue := s.Samples[0].Value
	var minIdx int
	for idx, sample := range s.Samples {
		if sample.Value < minValue {
			minValue = sample.Value
			minIdx = idx
		}
	}
	return s.Samples[minIdx], nil
}

func (s *TcmSamples) GetAvgPoint() (point *TcmSample, err error) {
	var sum float64
	for _, sample := range s.Samples {
		sum = sum + sample.Value
	}
	avg := sum / float64(len(s.Samples))
	sample := &TcmSample{
		Timestamp:  s.Samples[len(s.Samples)-1].Timestamp,
		Value:      avg,
		Dimensions: s.Samples[len(s.Samples)-1].Dimensions,
	}
	return sample, nil
}

func NewTcmSamples(series *TcmSeries, p *monitor.DataPoint) (s *TcmSamples, err error) {
	s = &TcmSamples{
		Series:  series,
		Samples: []*TcmSample{},
	}

	if len(p.Timestamps) == 0 {
		return nil, fmt.Errorf("Samples is empty ")
	}

	if len(p.Timestamps) != len(p.Values) {
		return nil, fmt.Errorf("Samples error, timestamps != values ")
	}

	for i := 0; i < len(p.Timestamps); i++ {
		s.Samples = append(s.Samples, &TcmSample{
			Timestamp:  *p.Timestamps[i],
			Value:      *p.Values[i],
			Dimensions: p.Dimensions,
		})
	}
	return
}
