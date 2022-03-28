package metrics

import (
	metrics1 "github.com/consensys/orchestrate/pkg/toolkit/app/metrics"
	pkgmetrics "github.com/consensys/orchestrate/pkg/toolkit/app/metrics/multi"
	"github.com/go-kit/kit/metrics/discard"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Subsystem = "transaction_listener"
	BlockName = "current_block"
)

type tpcMetrics struct {
	prometheus.Collector
	*metrics
}

func NewListenerMetrics() ListenerMetrics {
	multi := pkgmetrics.NewMulti()

	blockCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics1.Namespace,
			Subsystem: Subsystem,
			Name:      BlockName,
			Help:      "Current block processed",
		},
		[]string{"chain_uuid"},
	)
	multi.Collectors = append(multi.Collectors, blockCounter)

	return &tpcMetrics{
		Collector: multi,
		metrics: buildMetrics(
			kitprometheus.NewCounter(blockCounter),
		),
	}
}

func NewListenerNopMetrics() ListenerMetrics {
	return &tpcMetrics{
		Collector: pkgmetrics.NewMulti(),
		metrics: buildMetrics(
			discard.NewCounter(),
		),
	}
}
