// logging package provides a simple transport implementation for sending Prometheus metrics to a logging system.
package logging

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// LoggingTransport is a transport implementation that logs Prometheus metrics to a specified logger.
type LoggingTransport struct {
	logger *slog.Logger
}

// NewTransport creates a new instance of LoggingTransport with the provided logger.
func NewTransport(logger *slog.Logger) *LoggingTransport {
	return &LoggingTransport{
		logger: logger,
	}
}

func (lt LoggingTransport) Send(ctx context.Context, jobName string, data []prometheus.Collector) error {
	metrics := make(chan prometheus.Metric)
	go func() {
		defer close(metrics)
		for _, collector := range data {
			collector.Collect(metrics)
		}
	}()
	collected := make([]string, 0)
	for {
		select {
		case metric, ok := <-metrics:
			if !ok {
				lt.logger.InfoContext(
					ctx,
					"Sending data to logging transport",
					"job", jobName,
					"data", collected,
				)
				return nil
			}
			collected = append(collected, metric.Desc().String())
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (lt LoggingTransport) Validate() error {
	return nil
}
