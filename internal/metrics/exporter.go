package metrics

import (
	"time"

	"github.com/turgayozgur/messageman/internal/metrics/prometheus"
	"github.com/valyala/fasthttp"
)

// Exporter interface
type Exporter interface {
	Handle(ctx *fasthttp.RequestCtx)
	IncSendError(string, string)
	IncConsumeError(string, string)
	IncError(string)
	IncConsumer(string, string)
	DecConsumer(string, string)
	IncConnection(string)
	DecConnection(string)
	SendSeconds(time.Duration, string, string)
	ConsumeSeconds(time.Duration, string, string)
}

// CreateExporter factory method
func CreateExporter(enabled bool, name string) Exporter {
	if !enabled {
		return &NilExporter{}
	}
	switch name {
	case "prometheus":
		return prometheus.New()
	default:
		return prometheus.New()
	}
}

// NilExporter .
type NilExporter struct {
}

// Handler .
func (n *NilExporter) Handle(ctx *fasthttp.RequestCtx) {
}

// IncSendError .
func (n *NilExporter) IncSendError(service string, name string) {}

// IncConsumeError .
func (n *NilExporter) IncConsumeError(service string, name string) {}

// IncError .
func (n *NilExporter) IncError(service string) {}

// IncConsumer .
func (n *NilExporter) IncConsumer(service string, name string) {}

// DecConsumer .
func (n *NilExporter) DecConsumer(service string, name string) {}

// IncConnection .
func (n *NilExporter) IncConnection(name string) {}

// DecConnection .
func (n *NilExporter) DecConnection(name string) {}

// SendSeconds .
func (n *NilExporter) SendSeconds(d time.Duration, service string, name string) {}

// ConsumeSeconds .
func (n *NilExporter) ConsumeSeconds(d time.Duration, service string, name string) {}
