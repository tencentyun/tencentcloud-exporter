package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

// 负责一个指标的查询管理
type TcmQuery struct {
	Metric            *TcmMetric
	LatestQueryStatus int
	repo              TcmMetricRepository
}

type TcmQuerySet []*TcmQuery

func (q *TcmQuery) GetPromMetrics() (pms []prometheus.Metric, err error) {
	q.LatestQueryStatus = 2

	pms, err = q.Metric.GetLatestPromMetrics(q.repo)
	if err != nil {
		return
	}

	q.LatestQueryStatus = 1
	return
}

func (qs TcmQuerySet) SplitByBatch(batch int) (steps [][]*TcmQuery) {
	total := len(qs)
	for i := 0; i < total/batch+1; i++ {
		s := i * batch
		if s >= total {
			continue
		}
		e := i*batch + batch
		if e >= total {
			e = total
		}
		steps = append(steps, qs[s:e])
	}
	return
}

func NewTcmQuery(m *TcmMetric, repo TcmMetricRepository) (query *TcmQuery, err error) {
	query = &TcmQuery{
		Metric: m,
		repo:   repo,
	}
	return
}
