package transport

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

type Transport interface {
	Send(ctx context.Context, jobName string, data []prometheus.Collector) error
	Validate() error
}
