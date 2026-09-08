package core

import (
	"context"
	"time"

	"github.com/feronatech/go-aws-monitoring/monitoring/core/errors"
	"github.com/feronatech/go-aws-monitoring/monitoring/core/transport"
	"github.com/feronatech/go-aws-monitoring/monitoring/core/utils"

	"github.com/prometheus/client_golang/prometheus"
)

type metricType string

const (
	MonitoringContextKeyRegion      utils.MonitoringContextKey = "region"
	MonitoringContextKeyEnv         utils.MonitoringContextKey = "env"
	MonitoringContextKeyApplication utils.MonitoringContextKey = "application"

	MetricTypeHistogram metricType = "histogram"
	MetricTypeGauge     metricType = "gauge"
	MetricTypeCounter   metricType = "counter"
	MetricTypeSummary   metricType = "summary"
)

type MetricId struct {
	name  string
	mType metricType
}

type MonitoringOptions struct {
	Region      string
	Env         string
	Application string
	Transport   transport.Transport
}

type Monitoring interface {
	RegisterHistogram(name string, help string, unit string, buckets []float64) MetricId
	RegisterGauge(name string, help string, unit string) MetricId
	RegisterCounter(name string, help string, unit string) MetricId
	RegisterSummary(name string, help string, unit string, objectives map[float64]float64, maxAge time.Duration, ageBuckets uint32) MetricId
	Observe(id MetricId, value float64) error
	Add(id MetricId, value float64) error
	Set(id MetricId, value float64) error
	Sub(id MetricId, value float64) error
	Send(ctx context.Context, id MetricId) error
	Flush(ctx context.Context) error
}

type metricRegistry struct {
	histograms map[string]prometheus.Histogram
	gauges     map[string]prometheus.Gauge
	counters   map[string]prometheus.Counter
	summaries  map[string]prometheus.Summary
}

func (m *metricRegistry) getHistogram(id MetricId) (prometheus.Histogram, bool) {
	if id.mType != MetricTypeHistogram {
		return nil, false
	}
	hist, ok := m.histograms[id.name]
	return hist, ok
}

func (m *metricRegistry) getGauge(id MetricId) (prometheus.Gauge, bool) {
	if id.mType != MetricTypeGauge {
		return nil, false
	}
	gauge, ok := m.gauges[id.name]
	return gauge, ok
}

func (m *metricRegistry) getCounter(id MetricId) (prometheus.Counter, bool) {
	if id.mType != MetricTypeCounter {
		return nil, false
	}
	counter, ok := m.counters[id.name]
	return counter, ok
}

func (m *metricRegistry) getSummary(id MetricId) (prometheus.Summary, bool) {
	if id.mType != MetricTypeSummary {
		return nil, false
	}
	summary, ok := m.summaries[id.name]
	return summary, ok
}

type monitoring struct {
	region      string
	env         string
	application string
	transport   transport.Transport
	registry    *metricRegistry
}

func useOptionOrDefault(option string, defaultValue string) string {
	if option == "" {
		return defaultValue
	}
	return option
}

func SetupMonitoring(ctx context.Context, options MonitoringOptions) (Monitoring, error) {
	if err := options.Transport.Validate(); err != nil {
		return nil, errors.NewMonitoringConfigurationError("Transport Configuration invalid!", err)
	}

	return &monitoring{
		region:      useOptionOrDefault(options.Region, utils.GetValueFromContextAsStringOrDefault(ctx, MonitoringContextKeyRegion, "")),
		env:         useOptionOrDefault(options.Env, utils.GetValueFromContextAsStringOrDefault(ctx, MonitoringContextKeyEnv, "")),
		application: useOptionOrDefault(options.Application, utils.GetValueFromContextAsStringOrDefault(ctx, MonitoringContextKeyApplication, "")),
		transport:   options.Transport,
		registry: &metricRegistry{
			histograms: make(map[string]prometheus.Histogram),
			gauges:     make(map[string]prometheus.Gauge),
			counters:   make(map[string]prometheus.Counter),
			summaries:  make(map[string]prometheus.Summary),
		},
	}, nil
}

func (m *monitoring) RegisterHistogram(name string, help string, unit string, buckets []float64) MetricId {
	hist := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Unit:    unit,
		Buckets: buckets,
	})
	m.registry.histograms[name] = hist
	return MetricId{name: name, mType: MetricTypeHistogram}
}

func (m *monitoring) RegisterGauge(name string, help string, unit string) MetricId {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
		Unit: unit,
	})
	m.registry.gauges[name] = gauge
	return MetricId{name: name, mType: MetricTypeGauge}
}

func (m *monitoring) RegisterCounter(name string, help string, unit string) MetricId {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
		Unit: unit,
	})
	m.registry.counters[name] = counter
	return MetricId{name: name, mType: MetricTypeCounter}
}

func (m *monitoring) RegisterSummary(name string, help string, unit string, objectives map[float64]float64, maxAge time.Duration, ageBuckets uint32) MetricId {
	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       name,
		Help:       help,
		Unit:       unit,
		Objectives: objectives,
		MaxAge:     maxAge,
		AgeBuckets: ageBuckets,
	})
	m.registry.summaries[name] = summary
	return MetricId{name: name, mType: MetricTypeSummary}
}

func (m *monitoring) Observe(id MetricId, value float64) error {
	if hist, ok := m.registry.getHistogram(id); ok {
		hist.Observe(value)
		return nil
	}
	if summary, ok := m.registry.getSummary(id); ok {
		summary.Observe(value)
		return nil
	}
	return &errors.GeneralMonitoringError{Message: "Metric Id not found or not a histogram/summary!"}
}

func (m *monitoring) Add(id MetricId, value float64) error {
	if counter, ok := m.registry.getCounter(id); ok {
		counter.Add(value)
		return nil
	}
	if gauge, ok := m.registry.getGauge(id); ok {
		gauge.Add(value)
		return nil
	}
	return &errors.GeneralMonitoringError{Message: "Metric Id not found or not a counter/gauge!"}
}

func (m *monitoring) Set(id MetricId, value float64) error {
	if gauge, ok := m.registry.getGauge(id); ok {
		gauge.Set(value)
		return nil
	}
	return &errors.GeneralMonitoringError{Message: "Metric Id not found or not a gauge!"}
}

func (m *monitoring) Sub(id MetricId, value float64) error {
	if gauge, ok := m.registry.getGauge(id); ok {
		gauge.Sub(value)
		return nil
	}
	return &errors.GeneralMonitoringError{Message: "Metric Id not found or not a gauge!"}
}

func (m *monitoring) Send(ctx context.Context, id MetricId) error {
	if gauge, ok := m.registry.getGauge(id); ok {
		if err := m.transport.Send(ctx, m.application, []prometheus.Collector{gauge}); err != nil {
			return &errors.GeneralMonitoringError{Message: "Failed to send gauge metric!", Wrapped: err}
		}
		return nil
	}
	if counter, ok := m.registry.getCounter(id); ok {
		if err := m.transport.Send(ctx, m.application, []prometheus.Collector{counter}); err != nil {
			return &errors.GeneralMonitoringError{Message: "Failed to send counter metric!", Wrapped: err}
		}
		return nil
	}
	if hist, ok := m.registry.getHistogram(id); ok {
		if err := m.transport.Send(ctx, m.application, []prometheus.Collector{hist}); err != nil {
			return &errors.GeneralMonitoringError{Message: "Failed to send histogram metric!", Wrapped: err}
		}
		return nil
	}
	if summary, ok := m.registry.getSummary(id); ok {
		if err := m.transport.Send(ctx, m.application, []prometheus.Collector{summary}); err != nil {
			return &errors.GeneralMonitoringError{Message: "Failed to send summary metric!", Wrapped: err}
		}
		return nil
	}
	return &errors.GeneralMonitoringError{Message: "Metric Id not found!"}
}

func (m *monitoring) Flush(ctx context.Context) error {
	metrics := []prometheus.Collector{}
	for _, hist := range m.registry.histograms {
		metrics = append(metrics, hist)
	}
	for _, gauge := range m.registry.gauges {
		metrics = append(metrics, gauge)
	}
	for _, counter := range m.registry.counters {
		metrics = append(metrics, counter)
	}
	for _, summary := range m.registry.summaries {
		metrics = append(metrics, summary)
	}
	if err := m.transport.Send(ctx, m.application, metrics); err != nil {
		return &errors.GeneralMonitoringError{Message: "Failed to send metrics!", Wrapped: err}
	}
	return nil
}
