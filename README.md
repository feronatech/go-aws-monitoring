# Golang AWS Monitoring Utility Library

[![Static Badge](https://img.shields.io/badge/-v1.24-%23E6522C?logo=prometheus&logoColor=white&labelColor=gray)](go.mod) [![Static Badge](https://img.shields.io/badge/Go-1.25%2B-%2300ADD8?logo=go)](go.mod) [![Go Reference](https://img.shields.io/badge/Reference-%2300ADD8?logo=go&logoColor=white)](https://pkg.go.dev/github.com/feronatech/go-aws-monitoring) [![GitHub License](https://img.shields.io/github/license/feronatech/go-aws-monitoring?label=License&color=purple)](LICENSE)

A small Prometheus monitoring utility for applications running in AWS Lambda.

It provides a core monitoring interface for registering counters, gauges, histograms and summaries, then sends individual metrics or the complete registry through a transport. The included Pushgateway transport uses Prometheus' Pushgateway client and supports optional basic authentication. A logging transport is also provided for inspecting collected metric descriptors through a `log/slog` logger.

---

## Installation

```bash
go get github.com/feronatech/go-aws-monitoring
```

## Import

```go
import "github.com/feronatech/go-aws-monitoring/monitoring"
import "github.com/feronatech/go-aws-monitoring/monitoring/pushgateway"
```

For the logging transport:

```go
import "github.com/feronatech/go-aws-monitoring/monitoring/logging"
```

---

## Quick start

```go
package main

import (
    "context"
    "log/slog"

    "github.com/feronatech/go-aws-monitoring/monitoring"
    "github.com/feronatech/go-aws-monitoring/monitoring/pushgateway"
)

func main() {
    ctx := context.Background()

    transport := pushgateway.NewTransport(
        "http://pushgateway:9091",
        "username",
        "password",
    )

    metrics, err := monitoring.SetupMonitoring(ctx, monitoring.MonitoringOptions{
        Region:      "us-east-1",
        Env:         "prod",
        Application: "orders",
        Transport:   transport,
    })
    if err != nil {
        panic(err)
    }

    requests := metrics.RegisterCounter(
        "orders_requests_total",
        "Total number of order requests",
        "",
    )
    if err := metrics.Add(requests, 1); err != nil {
        panic(err)
    }
    if err := metrics.Flush(ctx); err != nil {
        panic(err)
    }

    _ = slog.Default() // The logging transport is available separately.
}
```

> `MetricId` values must be obtained from a registration method. Its identifying fields are unexported, so callers cannot construct one directly.

The usual flow is: create a transport, pass it to `monitoring.SetupMonitoring`, register a metric, update it with `Add`, `Set`, `Sub`, or `Observe`, and call `Flush` (or `Send` for one metric). `SetupMonitoring` validates the transport before returning.

---

## Packages

| Package | Import path | Purpose |
|---|---|---|
| `monitoring` | `github.com/feronatech/go-aws-monitoring/monitoring` | Public aliases and setup entry point. |
| `core` | `github.com/feronatech/go-aws-monitoring/monitoring/core` | Core implementation and API. |
| `pushgateway` | `github.com/feronatech/go-aws-monitoring/monitoring/pushgateway` | Push metrics to a Prometheus Pushgateway. |
| `logging` | `github.com/feronatech/go-aws-monitoring/monitoring/logging` | Log metric descriptors using `log/slog`. |

## Core API

### Exported constants and types

| Identifier | Declaration / signature | Description |
|---|---|---|
| `MonitoringContextKeyRegion` | `MonitoringContextKeyRegion utils.MonitoringContextKey = "region"` | Context key used as the fallback region. |
| `MonitoringContextKeyEnv` | `MonitoringContextKeyEnv utils.MonitoringContextKey = "env"` | Context key used as the fallback environment. |
| `MonitoringContextKeyApplication` | `MonitoringContextKeyApplication utils.MonitoringContextKey = "application"` | Context key used as the fallback application name and Pushgateway job name. |
| `MetricTypeHistogram` | `MetricTypeHistogram metricType = "histogram"` | Internal metric-kind value returned in a histogram `MetricId`. |
| `MetricTypeGauge` | `MetricTypeGauge metricType = "gauge"` | Internal metric-kind value returned in a gauge `MetricId`. |
| `MetricTypeCounter` | `MetricTypeCounter metricType = "counter"` | Internal metric-kind value returned in a counter `MetricId`. |
| `MetricTypeSummary` | `MetricTypeSummary metricType = "summary"` | Internal metric-kind value returned in a summary `MetricId`. |
| `MetricId` | `type MetricId struct { name string; mType metricType }` | Opaque handle identifying a registered metric. Both fields are unexported. |
| `MonitoringOptions` | `type MonitoringOptions struct { Region string; Env string; Application string; Transport transport.Transport }` | Setup configuration. Empty metadata fields fall back to context values. |
| `Monitoring` | `type Monitoring interface { ... }` | Registration, mutation, single-send and flush contract. |

The complete `Monitoring` interface is:

```go
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
```

### Setup

```go
func SetupMonitoring(ctx context.Context, options MonitoringOptions) (Monitoring, error)
```

Validates `options.Transport`. If validation fails, returns `nil` and a monitoring configuration error with message `Transport Configuration invalid!`, wrapping the transport error. On success, it creates an empty registry. For each of `Region`, `Env`, and `Application`, a non-empty option wins; otherwise the corresponding context key is read, with an empty-string default. No environment variables are read by the reviewed source.

### Registration and mutation methods

| Method | What it does |
|---|---|
| `RegisterHistogram(name string, help string, unit string, buckets []float64) MetricId` | Creates a Prometheus histogram with the supplied name, help, unit and buckets. |
| `RegisterGauge(name string, help string, unit string) MetricId` | Creates a Prometheus gauge. |
| `RegisterCounter(name string, help string, unit string) MetricId` | Creates a Prometheus counter. |
| `RegisterSummary(name string, help string, unit string, objectives map[float64]float64, maxAge time.Duration, ageBuckets uint32) MetricId` | Creates a Prometheus summary with the supplied objectives, age and age-bucket settings. |
| `Observe(id MetricId, value float64) error` | Observes a value on a histogram or summary. Otherwise returns `GeneralMonitoringError` with `Metric Id not found or not a histogram/summary!`. |
| `Add(id MetricId, value float64) error` | Adds to a counter or gauge. Otherwise returns `GeneralMonitoringError` with `Metric Id not found or not a counter/gauge!`. |
| `Set(id MetricId, value float64) error` | Sets a gauge. Otherwise returns `GeneralMonitoringError` with `Metric Id not found or not a gauge!`. |
| `Sub(id MetricId, value float64) error` | Subtracts from a gauge. Otherwise returns `GeneralMonitoringError` with `Metric Id not found or not a gauge!`. |
| `Send(ctx context.Context, id MetricId) error` | Sends exactly one registered collector through the transport. Transport errors are wrapped with a metric-specific message. |
| `Flush(ctx context.Context) error` | Sends all registered histograms, gauges, counters and summaries in registry-map iteration order. Transport errors are wrapped with `Failed to send metrics!`. |

All registered metrics receive constant labels named `region`, `environment`, and `application`. Duplicate names overwrite the corresponding internal map entry; the source does not report a duplicate-registration error.

> `Send` and `Flush` pass the monitoring `Application` value as the transport `jobName`. Context cancellation and transport-specific behavior are delegated to the transport.

---

## Pushgateway transport

```go
type PushgatewayTransport struct {
    url      string
    username string
    password string
}

func NewTransport(url string, username string, password string) *PushgatewayTransport
func (p PushgatewayTransport) Validate() error
func (p PushgatewayTransport) Send(ctx context.Context, jobName string, data []prometheus.Collector) error
```

`NewTransport` stores the URL and credentials. `Validate` rejects only an empty URL, returning a `TransportConfigurationError` with message `Pushgateway Url is required!`. `Send` builds a Prometheus pusher with `push.New(p.url, jobName)`, applies basic authentication only when both username and password are non-empty, adds every collector, and calls `PushContext(ctx)`.

> A username without a password, or a password without a username, does not enable basic authentication.

---

## Logging transport

```go
type LoggingTransport struct {
    logger *slog.Logger
}

func NewTransport(logger *slog.Logger) *LoggingTransport
func (lt LoggingTransport) Send(ctx context.Context, jobName string, data []prometheus.Collector) error
func (lt LoggingTransport) Validate() error
```

`NewTransport` stores the supplied logger. `Send` starts a goroutine, calls `Collect` on every supplied collector, and gathers each metric descriptor string (`metric.Desc().String()`). After collection it logs one `InfoContext` entry with message `Sending data to logging transport` and key/value pairs `job` and `data`. If `ctx` is canceled while waiting for metrics, it returns `ctx.Err()`. `Validate` always returns `nil`.

> The logging transport logs descriptor strings, not serialized metric samples. The source does not check for a nil logger; a nil logger can therefore panic when collection completes.

## License

See [LICENSE](LICENSE) for details.
