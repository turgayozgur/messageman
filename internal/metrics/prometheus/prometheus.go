package prometheus

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

var ()

// Prometheus .
type Prometheus struct {
	once                     sync.Once
	handlerFn                fasthttp.RequestHandler
	sentCounterVec           *prometheus.CounterVec
	receivedCounterVec       *prometheus.CounterVec
	retryCounterVec          *prometheus.CounterVec
	sendErrorCounterVec      *prometheus.CounterVec
	receiveErrorCounterVec   *prometheus.CounterVec
	errorCounterVec          *prometheus.CounterVec
	workerGaugeVec           *prometheus.GaugeVec
	connectionGaugeVec       *prometheus.GaugeVec
	sendDurationHistogram    *prometheus.HistogramVec
	receiveDurationHistogram *prometheus.HistogramVec
}

// New ctor
func New() *Prometheus {
	return &Prometheus{
		sendErrorCounterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "messageman_jobs_send_errors_total",
				Help: "Total number of send job errors",
			}, []string{"main_api", "queue"}),
		receiveErrorCounterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "messageman_jobs_receive_errors_total",
				Help: "Total number of receive job errors",
			}, []string{"main_api", "queue"}),
		errorCounterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "messageman_errors_total",
				Help: "Total number of errors except job errors",
			}, []string{"main_api"}),
		workerGaugeVec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "messageman_worker_connected_total",
				Help: "Number of active connected workers",
			}, []string{"main_api", "queue"}),
		connectionGaugeVec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "messageman_connection_active_total",
				Help: "Number of active connections",
			}, []string{"main_api"}),
		sendDurationHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "messageman_jobs_send_duration_seconds",
				Help:    "Send job duration seconds",
				Buckets: []float64{0.01, 0.1, 1, 5, 10},
			}, []string{"main_api", "queue"}),
		receiveDurationHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "messageman_jobs_receive_duration_seconds",
				Help:    "Receive job duration seconds",
				Buckets: []float64{0.01, 0.1, 1, 5, 10, 20, 60},
			}, []string{"main_api", "queue"}),
	}
}

// Handler the http handler for metrics.
func (p *Prometheus) Handle(ctx *fasthttp.RequestCtx) {
	p.once.Do(func() {
		r := prometheus.NewRegistry()

		// register ours.
		r.MustRegister(p.sendErrorCounterVec)
		r.MustRegister(p.receiveErrorCounterVec)
		r.MustRegister(p.errorCounterVec)
		r.MustRegister(p.workerGaugeVec)
		r.MustRegister(p.connectionGaugeVec)
		r.MustRegister(p.sendDurationHistogram)
		r.MustRegister(p.receiveDurationHistogram)

		p.handlerFn = fasthttpadaptor.NewFastHTTPHandler(promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	})

	p.handlerFn(ctx)
}

// IncSendError .
func (p *Prometheus) IncSendError(mainAPI string, queueName string) {
	p.sendErrorCounterVec.WithLabelValues(mainAPI, queueName).Inc()
}

// IncReceiveError .
func (p *Prometheus) IncReceiveError(mainAPI string, queueName string) {
	p.receiveErrorCounterVec.WithLabelValues(mainAPI, queueName).Inc()
}

// IncError .
func (p *Prometheus) IncError(mainAPI string) {
	p.errorCounterVec.WithLabelValues(mainAPI).Inc()
}

// IncWorker .
func (p *Prometheus) IncWorker(mainAPI string, queueName string) {
	p.workerGaugeVec.WithLabelValues(mainAPI, queueName).Inc()
}

// DecWorker .
func (p *Prometheus) DecWorker(mainAPI string, queueName string) {
	p.workerGaugeVec.WithLabelValues(mainAPI, queueName).Dec()
}

// IncConnection .
func (p *Prometheus) IncConnection(mainAPI string) {
	p.connectionGaugeVec.WithLabelValues(mainAPI).Inc()
}

// DecConnection .
func (p *Prometheus) DecConnection(mainAPI string) {
	p.connectionGaugeVec.WithLabelValues(mainAPI).Dec()
}

// SendSeconds .
func (p *Prometheus) SendSeconds(d time.Duration, mainAPI string, queueName string) {
	p.sendDurationHistogram.WithLabelValues(mainAPI, queueName).Observe(d.Seconds())
}

// ReceiveSeconds .
func (p *Prometheus) ReceiveSeconds(d time.Duration, mainAPI string, queueName string) {
	p.receiveDurationHistogram.WithLabelValues(mainAPI, queueName).Observe(d.Seconds())
}
