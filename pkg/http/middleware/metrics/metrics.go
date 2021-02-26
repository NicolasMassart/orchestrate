package metrics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ConsenSys/orchestrate/pkg/http/httputil"
	"github.com/ConsenSys/orchestrate/pkg/http/metrics"
	"github.com/ConsenSys/orchestrate/pkg/multitenancy"
)

type Builder struct {
	metrics metrics.HTTPMetrics
}

func NewBuilder(reg metrics.HTTPMetrics) *Builder {
	return &Builder{
		metrics: reg,
	}
}

func (b *Builder) Build(ctx context.Context, _ string, _ interface{}) (mid func(http.Handler) http.Handler, respModifier func(resp *http.Response) error, err error) {
	entrypoint := httputil.EntryPointFromContext(ctx)
	service := httputil.ServiceFromContext(ctx)

	m := New(
		b.metrics,
		[]string{"entrypoint", entrypoint, "service", service},
	)

	return m.Handler, nil, nil
}

type Metrics struct {
	registry   metrics.HTTPMetrics
	baseLabels []string
}

func New(registry metrics.HTTPMetrics, baseLabels []string) *Metrics {
	return &Metrics{
		registry:   registry,
		baseLabels: baseLabels,
	}
}

func (m *Metrics) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		m.ServeHTTP(rw, req, h)
	})
}

func (m *Metrics) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	authLabels := append(
		m.baseLabels,
		"tenant_id", multitenancy.TenantIDFromContext(req.Context()),
	)

	// Increment Conn Gauge
	connLabels := append(
		authLabels,
		"method", httputil.GetMethod(req),
		"protocol", httputil.GetProtocol(req),
	)

	openConnsGauge := m.registry.OpenConnsGauge().With(connLabels...)
	openConnsGauge.Add(1)
	defer openConnsGauge.Add(-1)

	recorder := httputil.NewResponseWriterRecorder(rw)
	start := time.Now()

	next.ServeHTTP(recorder, req)

	labels := append(connLabels, "code", strconv.Itoa(recorder.GetCode()))

	// Increment requests count
	m.registry.RequestsCounter().With(labels...).Add(1)

	if req.TLS != nil {
		tlsLabels := append(
			authLabels,
			"tls_version", httputil.GetTLSVersion(req),
			"tls_cipher", httputil.GetTLSCipher(req),
		)
		m.registry.TLSRequestsCounter().With(tlsLabels...).Add(1)
	}

	// Observe request latency
	d := float64(time.Since(start).Nanoseconds()) / float64(time.Second)
	if d < 0 {
		d = 0
	}
	m.registry.RequestsLatencyHistogram().With(labels...).Observe(d)
}
