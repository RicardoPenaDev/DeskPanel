package api

import (
	"context"

	"deskpanel-agent/internal/devices"
)

type contextKey int

const deviceContextKey contextKey = iota

func withDevice(ctx context.Context, d devices.Device) context.Context {
	return context.WithValue(ctx, deviceContextKey, d)
}

// DeviceFromContext retorna o dispositivo autenticado pelo middleware
// RequireBearerToken, se houver.
func DeviceFromContext(ctx context.Context) (devices.Device, bool) {
	d, ok := ctx.Value(deviceContextKey).(devices.Device)
	return d, ok
}
