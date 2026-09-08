package pushgateway

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func Test_NewTransport(t *testing.T) {
	transport := NewTransport("http://localhost:9091", "user", "pass")
	if transport == nil {
		t.Error("Expected non-nil transport")
	}
}

func Test_Validate(t *testing.T) {
	transport := NewTransport("", "user", "pass")
	err := transport.Validate()
	if err == nil {
		t.Error("Expected error for empty URL, got nil")
	}

	transport = NewTransport("http://localhost:9091", "user", "pass")
	err = transport.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid URL, got %v", err)
	}
}

func Test_Send(t *testing.T) {
	transport := NewTransport("http://localhost:9091", "user", "pass")

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
