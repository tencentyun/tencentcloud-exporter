package metric

import (
	"context"
	"fmt"
	"time"

	"github.com/tencentyun/tencentcloud-exporter/pkg/util"

	"github.com/tencentyun/tencentcloud-exporter/pkg/common"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/time/rate"

	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	v20180724 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	"github.com/tencentyun/tencentcloud-exporter/pkg/client"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
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
	credential               common.CredentialIface
	monitorClient            *monitor.Client
	monitorClientInGuangzhou *monitor.Client
	limiter                  *rate.Limiter // 限速
	ctx                      context.Context

	queryMetricBatchSize int

	logger log.Logger
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

	stStr := util.FormatTime(time.Unix(st, 0), timeStampFormat)
	request.StartTime = &stStr
	if et != 0 {
		etStr := util.FormatTime(time.Unix(et, 0), timeStampFormat)
		request.EndTime = &etStr
	}

	response := &v20180724.GetMonitorDataResponse{}
	if util.IsStrInList(config.QcloudNamespace, s.Metric.Meta.ProductName) {
		response, err = repo.monitorClientInGuangzhou.GetMonitorData(request)
	} else {
		response, err = repo.monitorClient.GetMonitorData(request)
	}
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

func (repo *TcmMetricRepositoryImpl) ListSamples(m *TcmMetric, st int64, et int64) ([]*TcmSamples, error) {
	var samplesList []*TcmSamples
	for _, seriesList := range m.GetSeriesSplitByBatch(repo.queryMetricBatchSize) {
		sl, err := repo.listSampleByBatch(m, seriesList, st, et)
		if err != nil {
			level.Error(repo.logger).Log("msg", err.Error())
			continue
		}
		samplesList = append(samplesList, sl...)
	}
	return samplesList, nil
}

func (repo *TcmMetricRepositoryImpl) listSampleByBatch(
	m *TcmMetric,
	seriesList []*TcmSeries,
	st int64,
	et int64,
) ([]*TcmSamples, error) {
	var samplesList []*TcmSamples

	ctx, cancel := context.WithCancel(repo.ctx)
	defer cancel()

	err := repo.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	request := repo.buildGetMonitorDataRequest(m, seriesList, st, et)

	response := &v20180724.GetMonitorDataResponse{}
	if util.IsStrInList(config.QcloudNamespace, m.Meta.ProductName) {
		response, err = repo.monitorClientInGuangzhou.GetMonitorData(request)
	} else {
		response, err = repo.monitorClient.GetMonitorData(request)
	}
	if err != nil {
		return nil, err
	}
	for _, points := range response.Response.DataPoints {
		samples, ql, e := repo.buildSamples(m, points)
		if e != nil {
			level.Debug(repo.logger).Log(
				"msg", e.Error(),
				"metric", m.Meta.MetricName,
				"dimension", fmt.Sprintf("%v", ql))
			continue
		}
		samplesList = append(samplesList, samples)
	}
	return samplesList, nil
}

func (repo *TcmMetricRepositoryImpl) buildGetMonitorDataRequest(
	m *TcmMetric,
	seriesList []*TcmSeries,
	st int64, et int64,
) *monitor.GetMonitorDataRequest {
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

	stStr := util.FormatTime(time.Unix(st, 0), timeStampFormat)
	request.StartTime = &stStr
	if et != 0 {
		etStr := util.FormatTime(time.Unix(et, 0), timeStampFormat)
		request.EndTime = &etStr
	}
	return request
}

func (repo *TcmMetricRepositoryImpl) buildSamples(
	m *TcmMetric,
	points *monitor.DataPoint,
) (*TcmSamples, map[string]string, error) {
	ql := map[string]string{}
	for _, dimension := range points.Dimensions {
		name := *dimension.Name
		if *dimension.Value != "" {
			_, ok := m.SeriesCache.LabelNames[name]
			if !ok {
				// if not in query label names, need ignore it
				// because series id = query labels md5
				continue
			}
			ql[name] = *dimension.Value
		}
	}
	sid, e := GetTcmSeriesId(m, ql)
	if e != nil {
		return nil, ql, fmt.Errorf("get series id fail")
	}
	s, ok := m.SeriesCache.Series[sid]
	if !ok {
		return nil, ql, fmt.Errorf("response data point not match series")
	}
	samples, e := NewTcmSamples(s, points)
	if e != nil {
		return nil, ql, fmt.Errorf("this instance may not have metric data")
	}
	return samples, ql, nil
}

func NewTcmMetricRepository(cred common.CredentialIface, conf *config.TencentConfig, logger log.Logger) (repo TcmMetricRepository, err error) {
	monitorClient, err := client.NewMonitorClient(cred, conf, conf.Credential.Region)
	if err != nil {
		return
	}
	monitorClientInGuangzhou, err := client.NewMonitorClient(cred, conf, "ap-guangzhou")
	if err != nil {
		return
	}

	repo = &TcmMetricRepositoryImpl{
		credential:               cred,
		monitorClient:            monitorClient,
		monitorClientInGuangzhou: monitorClientInGuangzhou,
		limiter:                  rate.NewLimiter(rate.Limit(conf.RateLimit), 1),
		ctx:                      context.Background(),
		queryMetricBatchSize:     conf.MetricQueryBatchSize,
		logger:                   logger,
	}

	return
}
