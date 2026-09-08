package logging

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func Test_NewTransport(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	transport := NewTransport(logger)
	if transport == nil {
		t.Error("Expected non-nil transport")
	}
}

func Test_Send(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	transport := NewTransport(logger)

	// Create a dummy prometheus.Collector for testing
	dummyGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "dummy_gauge",
		Help: "A dummy gauge for testing",
	})
	dummyGauge.Set(42)
	dummyCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dummy_counter",
		Help: "A dummy counter for testing",
	})
	dummyCounter.Add(3)
	dummyCounter.Add(2)

	err := transport.Send(context.Background(), "test_job", []prometheus.Collector{dummyGauge, dummyCounter})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
