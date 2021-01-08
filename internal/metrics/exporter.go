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
	IncReceiveError(string, string)
	IncPublishError(string, string)
	IncHandleError(string, string)
	IncError(string)
	IncWorker(string, string)
	DecWorker(string, string)
	IncSubscriber(string, string)
	DecSubscriber(string, string)
	IncConnection(string)
	DecConnection(string)
	SendSeconds(time.Duration, string, string)
	ReceiveSeconds(time.Duration, string, string)
	PublishSeconds(time.Duration, string, string)
	HandleSeconds(time.Duration, string, string)
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
func (n *NilExporter) IncSendError(mainAPI string, queueName string) {}

// IncReceiveError .
func (n *NilExporter) IncReceiveError(mainAPI string, queueName string) {}

// IncPublishError .
func (n *NilExporter) IncPublishError(publisher string, eventName string) {}

// IncHandleError .
func (n *NilExporter) IncHandleError(subscriber string, eventName string) {}

// IncError .
func (n *NilExporter) IncError(mainAPI string) {}

// IncWorker .
func (n *NilExporter) IncWorker(mainAPI string, queueName string) {}

// DecWorker .
func (n *NilExporter) DecWorker(mainAPI string, queueName string) {}

// IncSubscriber .
func (n *NilExporter) IncSubscriber(subscriber string, eventName string) {}

// DecSubscriber .
func (n *NilExporter) DecSubscriber(subscriber string, eventName string) {}

// IncConnection .
func (n *NilExporter) IncConnection(mainAPI string) {}

// DecConnection .
func (n *NilExporter) DecConnection(mainAPI string) {}

// SendSeconds .
func (n *NilExporter) SendSeconds(d time.Duration, mainAPI string, queueName string) {}

// ReceiveSeconds .
func (n *NilExporter) ReceiveSeconds(d time.Duration, mainAPI string, queueName string) {}

// PublishSeconds .
func (n *NilExporter) PublishSeconds(d time.Duration, mainAPI string, queueName string) {}

// HandleSeconds .
func (n *NilExporter) HandleSeconds(d time.Duration, mainAPI string, queueName string) {}
