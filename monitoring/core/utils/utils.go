package utils

import "context"

type MonitoringContextKey string

func GetValueFromContextAsStringOrDefault(ctx context.Context, key MonitoringContextKey, defaultValue string) string {
	if ctx == nil {
		return defaultValue
	}
	val := ctx.Value(key)
	if strVal, ok := val.(string); ok {
		return strVal
	}
	return defaultValue
}
