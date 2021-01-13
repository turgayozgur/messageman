package prometheus

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Prometheus .
type Prometheus struct {
	once                     sync.Once
	handlerFn                fasthttp.RequestHandler
	sendErrorCounterVec      *prometheus.CounterVec
	consumeErrorCounterVec   *prometheus.CounterVec
	errorCounterVec          *prometheus.CounterVec
	consumerGaugeVec         *prometheus.GaugeVec
	connectionGaugeVec       *prometheus.GaugeVec
	sendDurationHistogram    *prometheus.HistogramVec
	consumeDurationHistogram *prometheus.HistogramVec
}

// New ctor
func New() *Prometheus {
	return &Prometheus{
		sendErrorCounterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "messageman_send_errors_total",
				Help: "Total number of send errors",
			}, []string{"service", "name"}),
		consumeErrorCounterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "messageman_consume_errors_total",
				Help: "Total number of consume errors",
			}, []string{"service", "name"}),
		errorCounterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "messageman_errors_total",
				Help: "Total number of errors",
			}, []string{"service"}),
		consumerGaugeVec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "messageman_consumer_connected_total",
				Help: "Number of active connected consumers",
			}, []string{"service", "name"}),
		connectionGaugeVec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "messageman_connection_active_total",
				Help: "Number of active connections",
			}, []string{"name"}),
		sendDurationHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "messageman_send_duration_seconds",
				Help:    "Send duration seconds",
				Buckets: []float64{0.01, 0.1, 1, 5, 10},
			}, []string{"service", "name"}),
		consumeDurationHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "messageman_consume_duration_seconds",
				Help:    "Consume duration seconds",
				Buckets: []float64{0.01, 0.1, 1, 5, 10},
			}, []string{"service", "name"}),
	}
}

// Handler the http handler for metrics.
func (p *Prometheus) Handle(ctx *fasthttp.RequestCtx) {
	p.once.Do(func() {
		r := prometheus.NewRegistry()

		// register ours.
		r.MustRegister(p.sendErrorCounterVec)
		r.MustRegister(p.consumeErrorCounterVec)
		r.MustRegister(p.errorCounterVec)
		r.MustRegister(p.consumerGaugeVec)
		r.MustRegister(p.connectionGaugeVec)
		r.MustRegister(p.sendDurationHistogram)
		r.MustRegister(p.consumeDurationHistogram)

		p.handlerFn = fasthttpadaptor.NewFastHTTPHandler(promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	})

	p.handlerFn(ctx)
}

// IncSendError .
func (p *Prometheus) IncSendError(service string, name string) {
	p.sendErrorCounterVec.WithLabelValues(service, name).Inc()
}

// IncConsumeError .
func (p *Prometheus) IncConsumeError(service string, name string) {
	p.consumeErrorCounterVec.WithLabelValues(service, name).Inc()
}

// IncError .
func (p *Prometheus) IncError(service string) {
	p.errorCounterVec.WithLabelValues(service).Inc()
}

// IncConsumer .
func (p *Prometheus) IncConsumer(service string, name string) {
	p.consumerGaugeVec.WithLabelValues(service, name).Inc()
}

// DecConsumer .
func (p *Prometheus) DecConsumer(service string, name string) {
	p.consumerGaugeVec.WithLabelValues(service, name).Dec()
}

// IncConnection .
func (p *Prometheus) IncConnection(name string) {
	p.connectionGaugeVec.WithLabelValues(name).Inc()
}

// DecConnection .
func (p *Prometheus) DecConnection(name string) {
	p.connectionGaugeVec.WithLabelValues(name).Dec()
}

// SendSeconds .
func (p *Prometheus) SendSeconds(d time.Duration, service string, name string) {
	p.sendDurationHistogram.WithLabelValues(service, name).Observe(d.Seconds())
}

// ConsumeSeconds .
func (p *Prometheus) ConsumeSeconds(d time.Duration, service string, name string) {
	p.consumeDurationHistogram.WithLabelValues(service, name).Observe(d.Seconds())
}
