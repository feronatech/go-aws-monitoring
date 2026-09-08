// pushgateway package provides a transport implementation for sending Prometheus metrics to a Pushgateway server.
package pushgateway

import (
	"context"

	"github.com/feronatech/go-aws-monitoring/monitoring/core/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// PushgatewayTransport is a transport implementation that sends Prometheus metrics to a Pushgateway server.
// It supports basic authentication and requires a valid Pushgateway URL to function correctly.
type PushgatewayTransport struct {
	url      string
	username string
	password string
}

// NewTransport creates a new instance of PushgatewayTransport with the provided URL, username, and password.
func NewTransport(url string, username string, password string) *PushgatewayTransport {
	return &PushgatewayTransport{
		url:      url,
		username: username,
		password: password,
	}
}

func (p PushgatewayTransport) Validate() error {
	if p.url == "" {
		return &errors.TransportConfigurationError{Message: "Pushgateway Url is required!"}
	}
	return nil
}

func (p PushgatewayTransport) Send(ctx context.Context, jobName string, data []prometheus.Collector) error {
	pusher := push.New(p.url, jobName)
	if p.username != "" && p.password != "" {
		pusher.BasicAuth(p.username, p.password)
	}
	for _, c := range data {
		pusher.Collector(c)
	}
	return pusher.PushContext(ctx)
}
