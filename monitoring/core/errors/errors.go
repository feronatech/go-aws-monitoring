package errors

type MonitoringConfigurationError struct {
	Message string
	Wrapped error
}

func (e *MonitoringConfigurationError) Error() string {
	return e.Message
}

func (e *MonitoringConfigurationError) Unwrap() error {
	return e.Wrapped
}

func NewMonitoringConfigurationError(message string, wrapped error) *MonitoringConfigurationError {
	return &MonitoringConfigurationError{
		Message: message,
		Wrapped: wrapped,
	}
}

type TransportConfigurationError struct {
	Message string
}

func (e *TransportConfigurationError) Error() string {
	return e.Message
}

type GeneralMonitoringError struct {
	Message string
	Wrapped error
}

func (e *GeneralMonitoringError) Error() string {
	return e.Message
}
