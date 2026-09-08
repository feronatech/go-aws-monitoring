package utils

import (
	"context"
	"testing"
)

func Test_GetValueFromContextAsStringOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		key          MonitoringContextKey
		defaultValue string
		want         string
	}{
		{
			name:         "value present",
			ctx:          context.WithValue(context.Background(), MonitoringContextKey("test"), "value"),
			key:          MonitoringContextKey("test"),
			defaultValue: "default",
			want:         "value",
		},
		{
			name:         "value not present",
			ctx:          context.Background(),
			key:          MonitoringContextKey("test"),
			defaultValue: "default",
			want:         "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetValueFromContextAsStringOrDefault(tt.ctx, tt.key, tt.defaultValue); got != tt.want {
				t.Errorf("GetValueFromContextAsStringOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
