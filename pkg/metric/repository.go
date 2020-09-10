package metric

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
	"golang.org/x/time/rate"
	"time"
)

var (
	timeStampFormat = "2006-01-02 15:04:05"
)

// 腾讯云监控指标Repository
type TcmMetricRepository interface {
	// 获取指标的元数据
	GetMeta(namespace string, name string) (*TcmMeta, error)
	// 根据namespace获取所有的指标元数据
	ListMetaByNamespace(namespace string) ([]*TcmMeta, error)
	// 按时间范围获取单个时间线的数据点
	GetSamples(series *TcmSeries, startTime int64, endTime int64) (samples *TcmSamples, err error)
	// 按时间范围获取单个指标下所有时间线的数据点
	ListSamples(metric *TcmMetric, startTime int64, endTime int64) (samplesList []*TcmSamples, err error)
}

type TcmMetricRepositoryImpl struct {
	monitorClient *monitor.Client
	limiter       *rate.Limiter // 限速
	ctx           context.Context
	logger        log.Logger
}

func (repo *TcmMetricRepositoryImpl) GetMeta(namespace string, name string) (meta *TcmMeta, err error) {
	// 限速
	ctx, cancel := context.WithCancel(repo.ctx)
	defer cancel()
	err = repo.limiter.Wait(ctx)
	if err != nil {
		return
	}

	request := monitor.NewDescribeBaseMetricsRequest()
	request.Namespace = &namespace
	request.MetricName = &name
	response, err := repo.monitorClient.DescribeBaseMetrics(request)
	if err != nil {
		return
	}
	if len(response.Response.MetricSet) != 1 {
		return nil, fmt.Errorf("response metricSet size != 1")
	}
	meta, err = NewTcmMeta(response.Response.MetricSet[0])
	if err != nil {
		return
	}
	return
}

func (repo *TcmMetricRepositoryImpl) ListMetaByNamespace(namespace string) (metas []*TcmMeta, err error) {
	// 限速
	ctx, cancel := context.WithCancel(repo.ctx)
	defer cancel()
	err = repo.limiter.Wait(ctx)
	if err != nil {
		return
	}

	request := monitor.NewDescribeBaseMetricsRequest()
	request.Namespace = &namespace
	response, err := repo.monitorClient.DescribeBaseMetrics(request)
	if err != nil {
		return
	}
	for _, metricSet := range response.Response.MetricSet {
		m, e := NewTcmMeta(metricSet)
		if e != nil {
			return nil, err
		}
		metas = append(metas, m)
	}
	return
}

func (repo *TcmMetricRepositoryImpl) GetSamples(s *TcmSeries, st int64, et int64) (samples *TcmSamples, err error) {
	// 限速
	ctx, cancel := context.WithCancel(repo.ctx)
	defer cancel()
	err = repo.limiter.Wait(ctx)
	if err != nil {
		return
	}

	request := monitor.NewGetMonitorDataRequest()
	request.Namespace = &s.Metric.Meta.Namespace
	request.MetricName = &s.Metric.Meta.MetricName

	period := uint64(s.Metric.Conf.StatPeriodSeconds)
	request.Period = &period

	instanceFilters := &monitor.Instance{
		Dimensions: []*monitor.Dimension{},
	}
	for k, v := range s.QueryLabels {
		tk := k
		tv := v
		instanceFilters.Dimensions = append(instanceFilters.Dimensions, &monitor.Dimension{Name: &tk, Value: &tv})
	}
	request.Instances = []*monitor.Instance{instanceFilters}

	stStr := time.Unix(st, 0).Format(timeStampFormat)
	request.StartTime = &stStr
	if et != 0 {
		etStr := time.Unix(et, 0).Format(timeStampFormat)
		request.StartTime = &etStr
	}

	response, err := repo.monitorClient.GetMonitorData(request)
	if err != nil {
		return
	}

	if len(response.Response.DataPoints) != 1 {
		return nil, fmt.Errorf("response dataPoints size!=1")
	}

	samples, err = NewTcmSamples(s, response.Response.DataPoints[0])
	if err != nil {
		return
	}
	return
}

func (repo *TcmMetricRepositoryImpl) ListSamples(m *TcmMetric, st int64, et int64) (samplesList []*TcmSamples, err error) {
	for _, seriesList := range m.GetSeriesSplitByBatch(10) {
		ctx, cancel := context.WithCancel(repo.ctx)
		err = repo.limiter.Wait(ctx)
		if err != nil {
			return
		}

		request := monitor.NewGetMonitorDataRequest()
		request.Namespace = &m.Meta.Namespace
		request.MetricName = &m.Meta.MetricName

		period := uint64(m.Conf.StatPeriodSeconds)
		request.Period = &period

		for _, series := range seriesList {
			ifilters := &monitor.Instance{
				Dimensions: []*monitor.Dimension{},
			}
			for k, v := range series.QueryLabels {
				tk := k
				tv := v
				ifilters.Dimensions = append(ifilters.Dimensions, &monitor.Dimension{Name: &tk, Value: &tv})
			}
			request.Instances = append(request.Instances, ifilters)
		}

		stStr := time.Unix(st, 0).Format(timeStampFormat)
		request.StartTime = &stStr
		if et != 0 {
			etStr := time.Unix(et, 0).Format(timeStampFormat)
			request.StartTime = &etStr
		}

		response, err := repo.monitorClient.GetMonitorData(request)
		if err != nil {
			return nil, err
		}

		for _, points := range response.Response.DataPoints {
			ql := map[string]string{}
			for _, dimension := range points.Dimensions {
				if *dimension.Value != "" {
					ql[*dimension.Name] = *dimension.Value
				}
			}
			sid, e := GetTcmSeriesId(m, ql)
			if e != nil {
				level.Warn(repo.logger).Log(
					"msg", "Get series id fail",
					"metric", m.Meta.MetricName,
					"dimension", fmt.Sprintf("%v", ql))
				continue
			}
			s, ok := m.Series[sid]
			if !ok {
				level.Warn(repo.logger).Log(
					"msg", "Response data point not match series",
					"metric", m.Meta.MetricName,
					"dimension", fmt.Sprintf("%v", ql))
				continue
			}
			samples, e := NewTcmSamples(s, points)
			if e != nil {
				level.Warn(repo.logger).Log(
					"msg", "Instance has not monitor data",
					"metric", m.Meta.MetricName,
					"dimension", fmt.Sprintf("%v", ql))
			} else {
				samplesList = append(samplesList, samples)
			}

		}

		cancel()
	}
	return
}

func NewTcmMetricRepository(conf *config.TencentConfig, logger log.Logger) (repo TcmMetricRepository, err error) {
	monitorClient, err := client.NewMonitorClient(conf)
	if err != nil {
		return
	}

	repo = &TcmMetricRepositoryImpl{
		monitorClient: monitorClient,
		limiter:       rate.NewLimiter(rate.Limit(conf.RateLimit), 1),
		ctx:           context.Background(),
		logger:        logger,
	}

	return
}
