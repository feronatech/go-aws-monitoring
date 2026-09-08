package core

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type testCollector struct {
	Collectors []prometheus.Collector
}

type TestTransport struct {
	collector *testCollector
}

func (t TestTransport) Send(ctx context.Context, jobName string, data []prometheus.Collector) error {
	log.Printf("TestTransport: Sending data for job %s with %d collectors", jobName, len(data))
	t.collector.Collectors = append(t.collector.Collectors, data...)
	return nil
}

func (t TestTransport) Validate() error {
	return nil
}

func Test_SetupMonitoring(t *testing.T) {
	ctx := context.Background()
	transport := TestTransport{collector: &testCollector{}}
	region := "us-west-1"
	env := "production"
	app := "myapp"

	options := MonitoringOptions{
		Region:      region,
		Env:         env,
		Application: app,
		Transport:   transport,
	}
	m, err := SetupMonitoring(ctx, options)
	if err != nil {
		t.Errorf("Failed to setup monitoring: %v", err)
	}
	if m == nil {
		t.Errorf("Expected monitoring instance, got nil")
	}
	monitoringImpl, ok := m.(*monitoring)
	if !ok {
		t.Errorf("Expected monitoring to be of type *monitoring, got %T", m)
	}
	if monitoringImpl.region != region {
		t.Errorf("Expected region to be %s, got %s", region, monitoringImpl.region)
	}
	if monitoringImpl.env != env {
		t.Errorf("Expected env to be %s, got %s", env, monitoringImpl.env)
	}
	if monitoringImpl.application != app {
		t.Errorf("Expected application to be %s, got %s", app, monitoringImpl.application)
	}
	if monitoringImpl.registry == nil {
		t.Errorf("Expected registry to be initialized, got nil")
	}
	if monitoringImpl.registry.counters == nil || monitoringImpl.registry.gauges == nil || monitoringImpl.registry.histograms == nil || monitoringImpl.registry.summaries == nil {
		t.Errorf("Expected registry to be initialized, got nil")
	}
}

func Test_RegisterHistogram(t *testing.T) {
	ctx := context.Background()
	transport := TestTransport{collector: &testCollector{}}
	m, err := SetupMonitoring(ctx, MonitoringOptions{
		Region:      "us-west-1",
		Env:         "production",
		Application: "myapp",
		Transport:   transport,
	})
	if err != nil {
		t.Errorf("Failed to setup monitoring: %v", err)
	}

	name := "test_histogram"
	help := "This is a test histogram"
	unit := "seconds"
	buckets := []float64{0.1, 0.5, 1.0, 5.0, 10.0}
	id := m.RegisterHistogram(name, help, unit, buckets)
	if id.name != name {
		t.Errorf("Expected histogram name to be %s, got %s", name, id.name)
	}
	if id.mType != MetricTypeHistogram {
		t.Errorf("Expected metric type to be Histogram, got %v", id.mType)
	}

	monitoringImpl, _ := m.(*monitoring)
	if _, ok := monitoringImpl.registry.getHistogram(id); !ok {
		t.Errorf("Expected histogram to be registered in the registry")
	}
}

func Test_RegisterCounter(t *testing.T) {
	ctx := context.Background()
	transport := TestTransport{collector: &testCollector{}}
	m, err := SetupMonitoring(ctx, MonitoringOptions{
		Region:      "us-west-1",
		Env:         "production",
		Application: "myapp",
		Transport:   transport,
	})
	if err != nil {
		t.Errorf("Failed to setup monitoring: %v", err)
	}

	name := "test_counter"
	help := "This is a test counter"
	unit := "count"
	id := m.RegisterCounter(name, help, unit)
	if id.name != name {
		t.Errorf("Expected counter name to be %s, got %s", name, id.name)
	}
	if id.mType != MetricTypeCounter {
		t.Errorf("Expected metric type to be Counter, got %v", id.mType)
	}

	monitoringImpl, _ := m.(*monitoring)
	if _, ok := monitoringImpl.registry.getCounter(id); !ok {
		t.Errorf("Expected counter to be registered in the registry")
	}
}

func Test_RegisterGauge(t *testing.T) {
	ctx := context.Background()
	transport := TestTransport{collector: &testCollector{}}
	m, err := SetupMonitoring(ctx, MonitoringOptions{
		Region:      "us-west-1",
		Env:         "production",
		Application: "myapp",
		Transport:   transport,
	})
	if err != nil {
		t.Errorf("Failed to setup monitoring: %v", err)
	}

	name := "test_gauge"
	help := "This is a test gauge"
	unit := "count"
	id := m.RegisterGauge(name, help, unit)
	if id.name != name {
		t.Errorf("Expected gauge name to be %s, got %s", name, id.name)
	}
	if id.mType != MetricTypeGauge {
		t.Errorf("Expected metric type to be Gauge, got %v", id.mType)
	}

	monitoringImpl, _ := m.(*monitoring)
	if _, ok := monitoringImpl.registry.getGauge(id); !ok {
		t.Errorf("Expected gauge to be registered in the registry")
	}
}

func Test_RegisterSummary(t *testing.T) {
	ctx := context.Background()
	transport := TestTransport{collector: &testCollector{}}
	m, err := SetupMonitoring(ctx, MonitoringOptions{
		Region:      "us-west-1",
		Env:         "production",
		Application: "myapp",
		Transport:   transport,
	})
	if err != nil {
		t.Errorf("Failed to setup monitoring: %v", err)
	}

	name := "test_summary"
	help := "This is a test summary"
	unit := "seconds"
	objectives := map[float64]float64{0.5: 0.5, 0.9: 0.9, 0.99: 0.99}
	id := m.RegisterSummary(name, help, unit, objectives, 10*time.Minute, 5)
	if id.name != name {
		t.Errorf("Expected summary name to be %s, got %s", name, id.name)
	}
	if id.mType != MetricTypeSummary {
		t.Errorf("Expected metric type to be Summary, got %v", id.mType)
	}

	monitoringImpl, _ := m.(*monitoring)
	if _, ok := monitoringImpl.registry.getSummary(id); !ok {
		t.Errorf("Expected summary to be registered in the registry")
	}
}

const (
	TestMetricHistogramName = "test_histogram"
	TestMetricGaugeName     = "test_gauge"
	TestMetricCounterName   = "test_counter"
	TestMetricSummaryName   = "test_summary"
)

func setupTestMonitoring(t *testing.T) (Monitoring, *TestTransport) {
	t.Helper()
	ctx := context.Background()
	transport := TestTransport{collector: &testCollector{}}
	m, _ := SetupMonitoring(ctx, MonitoringOptions{
		Region:      "us-west-1",
		Env:         "production",
		Application: "myapp",
		Transport:   transport,
	})
	m.RegisterHistogram(TestMetricHistogramName, "Histogram for testing", "seconds", []float64{0, 0.333, 0.666, 1.0})
	m.RegisterGauge(TestMetricGaugeName, "Testing gauge", "kpi")
	m.RegisterCounter(TestMetricCounterName, "Counter for testing", "requests")
	m.RegisterSummary(TestMetricSummaryName, "Testing summary", "seconds", map[float64]float64{0.5: 0.5, 0.9: 0.9, 0.99: 0.99}, 10*time.Minute, 5)
	return m, &transport
}

func Test_Observe(t *testing.T) {
	m, _ := setupTestMonitoring(t)
	testCases := []struct {
		mName         string
		mType         metricType
		value         float64
		expectedError bool
	}{
		{TestMetricHistogramName, MetricTypeHistogram, 0.5, false},
		{TestMetricHistogramName, MetricTypeGauge, 0.5, true},
		{TestMetricGaugeName, MetricTypeGauge, 1.0, true},
		{TestMetricCounterName, MetricTypeCounter, 1.0, true},
		{TestMetricSummaryName, MetricTypeSummary, 1.0, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s = %t", tc.mName, tc.mType, tc.expectedError), func(t *testing.T) {
			id := MetricId{name: tc.mName, mType: tc.mType}
			err := m.Observe(id, tc.value)
			if (err != nil) != tc.expectedError {
				t.Errorf("Observe() error = %v, expectedError %v", err, tc.expectedError)
			}
		})
	}
}

func Test_Add(t *testing.T) {
	m, _ := setupTestMonitoring(t)
	testCases := []struct {
		mName         string
		mType         metricType
		value         float64
		expectedError bool
	}{
		{TestMetricHistogramName, MetricTypeHistogram, 0.5, true},
		{TestMetricGaugeName, MetricTypeGauge, 1.0, false},
		{TestMetricCounterName, MetricTypeCounter, 1.0, false},
		{TestMetricSummaryName, MetricTypeSummary, 1.0, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s = %t", tc.mName, tc.mType, tc.expectedError), func(t *testing.T) {
			id := MetricId{name: tc.mName, mType: tc.mType}
			err := m.Add(id, tc.value)
			if (err != nil) != tc.expectedError {
				t.Errorf("Add() error = %v, expectedError %v", err, tc.expectedError)
			}
		})
	}
}

func Test_Set(t *testing.T) {
	m, _ := setupTestMonitoring(t)
	testCases := []struct {
		mName         string
		mType         metricType
		value         float64
		expectedError bool
	}{
		{TestMetricHistogramName, MetricTypeHistogram, 0.5, true},
		{TestMetricGaugeName, MetricTypeGauge, 1.0, false},
		{TestMetricGaugeName, MetricTypeCounter, 1.0, true},
		{TestMetricCounterName, MetricTypeCounter, 1.0, true},
		{TestMetricSummaryName, MetricTypeSummary, 1.0, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s = %t", tc.mName, tc.mType, tc.expectedError), func(t *testing.T) {
			id := MetricId{name: tc.mName, mType: tc.mType}
			err := m.Set(id, tc.value)
			if (err != nil) != tc.expectedError {
				t.Errorf("Set() error = %v, expectedError %v", err, tc.expectedError)
			}
		})
	}
}

func Test_Sub(t *testing.T) {
	m, _ := setupTestMonitoring(t)
	testCases := []struct {
		mName         string
		mType         metricType
		value         float64
		expectedError bool
	}{
		{TestMetricHistogramName, MetricTypeHistogram, 0.5, true},
		{TestMetricGaugeName, MetricTypeGauge, 1.0, false},
		{TestMetricGaugeName, MetricTypeCounter, 1.0, true},
		{TestMetricCounterName, MetricTypeCounter, 1.0, true},
		{TestMetricSummaryName, MetricTypeSummary, 1.0, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s = %t", tc.mName, tc.mType, tc.expectedError), func(t *testing.T) {
			id := MetricId{name: tc.mName, mType: tc.mType}
			err := m.Sub(id, tc.value)
			if (err != nil) != tc.expectedError {
				t.Errorf("Sub() error = %v, expectedError %v", err, tc.expectedError)
			}
		})
	}
}

func Test_Send(t *testing.T) {
	testCases := []struct {
		mName         string
		mType         metricType
		value         float64
		expectedError bool
	}{
		{TestMetricHistogramName, MetricTypeHistogram, 0.5, false},
		{TestMetricHistogramName, MetricTypeGauge, 0.5, true},
		{TestMetricGaugeName, MetricTypeGauge, 1.0, false},
		{TestMetricGaugeName, MetricTypeCounter, 1.0, true},
		{TestMetricCounterName, MetricTypeCounter, 1.0, false},
		{TestMetricCounterName, MetricTypeSummary, 1.0, true},
		{TestMetricSummaryName, MetricTypeSummary, 1.0, false},
		{TestMetricSummaryName, MetricTypeHistogram, 1.0, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s = %t", tc.mName, tc.mType, tc.expectedError), func(t *testing.T) {
			m, transport := setupTestMonitoring(t)
			id := MetricId{name: tc.mName, mType: tc.mType}
			if tc.mName == TestMetricHistogramName || tc.mName == TestMetricSummaryName {
				_ = m.Observe(id, tc.value)
			} else {
				_ = m.Add(id, tc.value)
			}
			err := m.Send(t.Context(), id)
			if (err != nil) != tc.expectedError {
				t.Errorf("Send() error = %v, expectedError %v", err, tc.expectedError)
			}
			if !tc.expectedError && len(transport.collector.Collectors) == 0 {
				t.Errorf("Expected metrics to be sent, but transport has no collectors")
			}
		})
	}
}

func Test_Flush(t *testing.T) {
	testData := []struct {
		mName string
		mType metricType
		value float64
	}{
		{TestMetricHistogramName, MetricTypeHistogram, 0.5},
		{TestMetricHistogramName, MetricTypeGauge, 0.5},
		{TestMetricGaugeName, MetricTypeGauge, 1.0},
		{TestMetricGaugeName, MetricTypeCounter, 1.0},
		{TestMetricCounterName, MetricTypeCounter, 1.0},
		{TestMetricCounterName, MetricTypeSummary, 1.0},
		{TestMetricSummaryName, MetricTypeSummary, 1.0},
		{TestMetricSummaryName, MetricTypeHistogram, 1.0},
	}

	m, transport := setupTestMonitoring(t)
	for _, data := range testData {
		id := MetricId{name: data.mName, mType: data.mType}
		if data.mName == TestMetricHistogramName || data.mName == TestMetricSummaryName {
			_ = m.Observe(id, data.value)
		} else {
			_ = m.Add(id, data.value)
		}
	}

	err := m.Flush(t.Context())
	if err != nil {
		t.Errorf("Flush() error = %v, expected no error", err)
	}
	if len(transport.collector.Collectors) != 4 {
		t.Errorf("Expected 4 metrics to be sent, but transport has no collectors")
	}
}
