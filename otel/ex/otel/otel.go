package otel

import (
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
)

const otelName = "utils"

var (
	Tracer = otel.Tracer(otelName)
	Start  = Tracer.Start
)

var (
	Log = otelslog.NewLogger(otelName)

	With  = Log.With
	Info  = Log.InfoContext
	Error = Log.ErrorContext
)

var Meter = otel.Meter(otelName)
