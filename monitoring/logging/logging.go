package logging

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

type LoggingTransport struct {
	logger *slog.Logger
}

func NewTransport(logger *slog.Logger) *LoggingTransport {
	return &LoggingTransport{
		logger: logger,
	}
}

func (lt LoggingTransport) Send(ctx context.Context, jobName string, data []prometheus.Collector) error {
	metric := make(chan prometheus.Metric)
	for _, c := range data {
		c.Collect(metric)
	}
	lt.logger.Info("Sending data to logging transport", "job", jobName, "data", metric)
	return nil
}

func (lt LoggingTransport) Validate() error {
	return nil
}
