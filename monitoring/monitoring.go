// monitoring package provides a simple interface for monitoring applications using Prometheus metrics.
// It allows you to create and manage different types of metrics such as counters, gauges, histograms, and summaries.
// The package also provides functionality to send these metrics to a specified transport for further processing or storage.
package monitoring

import "github.com/feronatech/go-aws-monitoring/monitoring/core"

// MetricId is a type alias for core.MetricId, representing the unique identifier for a Prometheus metric.
// Do not create it by yourself! It is returned when registering a metric and is used to update the metric later on.
type MetricId = core.MetricId

// MetricLabel is a type alias for core.MetricLabel, representing a key-value pair used to label Prometheus metrics.
// It can be provided when updateing a metric to add additional context or metadata to the metric.
type MetricLLabel = core.MetricLabel

// Monitoring is an interface that defines methods for registering and managing Prometheus metrics, as well as sending them to a specified transport.
type Monitoring = core.Monitoring

// MonitoringOptions is a configuration object used to initialize a Monitoring instance.
// It includes fields for region, environment, application name, and the transport to be used for sending metrics.
// If certain fields are not provided in the MonitoringOptions, it will try to retrieve the information from the provided context.
// For concrete transport implementations, see the logging and pushgateway packages.
type MonitoringOptions = core.MonitoringOptions

// SetupMonitoring is a function that initializes and returns a Monitoring instance based on the provided MonitoringOptions.
var SetupMonitoring = core.SetupMonitoring
