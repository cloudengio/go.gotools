package apilog

import "context"

func LogCallfLegacy(ctx *context.Context, format string, v ...interface{}) func(*context.Context, string, ...interface{}) {
	return nil
}

func LogCallf(ctx *context.Context, name, callerLocation, format string, v ...interface{}) func(*context.Context, string, ...interface{}) {
	return nil
}
