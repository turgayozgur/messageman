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
	retryCounterVec          *prometheus.CounterVec
	sendErrorCounterVec      *prometheus.CounterVec
	receiveErrorCounterVec   *prometheus.CounterVec
	publishErrorCounterVec   *prometheus.CounterVec
	handleErrorCounterVec    *prometheus.CounterVec
	errorCounterVec          *prometheus.CounterVec
	workerGaugeVec           *prometheus.GaugeVec
	subscriberGaugeVec       *prometheus.GaugeVec
	connectionGaugeVec       *prometheus.GaugeVec
	sendDurationHistogram    *prometheus.HistogramVec
	receiveDurationHistogram *prometheus.HistogramVec
	publishDurationHistogram *prometheus.HistogramVec
	handleDurationHistogram  *prometheus.HistogramVec
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
		publishErrorCounterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "messageman_publish_errors_total",
				Help: "Total number of publish errors",
			}, []string{"publisher", "event_name"}),
		handleErrorCounterVec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "messageman_handle_errors_total",
				Help: "Total number of handle errors",
			}, []string{"subscriber", "event_name"}),
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
		subscriberGaugeVec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "messageman_subscriber_connected_total",
				Help: "Number of active connected subscribers",
			}, []string{"subscriber", "event_name"}),
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
		publishDurationHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "messageman_publish_duration_seconds",
				Help:    "Publish duration seconds",
				Buckets: []float64{0.01, 0.1, 1, 5, 10},
			}, []string{"publisher", "event_name"}),
		handleDurationHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "messageman_handle_duration_seconds",
				Help:    "Handle duration seconds",
				Buckets: []float64{0.01, 0.1, 1, 5, 10, 20, 60},
			}, []string{"subscriber", "event_name"}),
	}
}

// Handler the http handler for metrics.
func (p *Prometheus) Handle(ctx *fasthttp.RequestCtx) {
	p.once.Do(func() {
		r := prometheus.NewRegistry()

		// register ours.
		r.MustRegister(p.sendErrorCounterVec)
		r.MustRegister(p.receiveErrorCounterVec)
		r.MustRegister(p.publishErrorCounterVec)
		r.MustRegister(p.handleErrorCounterVec)
		r.MustRegister(p.errorCounterVec)
		r.MustRegister(p.workerGaugeVec)
		r.MustRegister(p.subscriberGaugeVec)
		r.MustRegister(p.connectionGaugeVec)
		r.MustRegister(p.sendDurationHistogram)
		r.MustRegister(p.receiveDurationHistogram)
		r.MustRegister(p.publishDurationHistogram)
		r.MustRegister(p.handleDurationHistogram)

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

// IncPublishError .
func (p *Prometheus) IncPublishError(publisher string, eventName string) {
	p.publishErrorCounterVec.WithLabelValues(publisher, eventName).Inc()
}

// IncHandleError .
func (p *Prometheus) IncHandleError(subscriber string, eventName string) {
	p.subscriberGaugeVec.WithLabelValues(subscriber, eventName).Inc()
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

// IncSubscriber .
func (p *Prometheus) IncSubscriber(subscriber string, eventName string) {
	p.subscriberGaugeVec.WithLabelValues(subscriber, eventName).Inc()
}

// DecSubscriber .
func (p *Prometheus) DecSubscriber(subscriber string, eventName string) {
	p.subscriberGaugeVec.WithLabelValues(subscriber, eventName).Dec()
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

// PublishSeconds .
func (p *Prometheus) PublishSeconds(d time.Duration, publisher string, eventName string) {
	p.publishDurationHistogram.WithLabelValues(publisher, eventName).Observe(d.Seconds())
}

// HandleSeconds .
func (p *Prometheus) HandleSeconds(d time.Duration, subscriber string, eventName string) {
	p.handleDurationHistogram.WithLabelValues(subscriber, eventName).Observe(d.Seconds())
}
