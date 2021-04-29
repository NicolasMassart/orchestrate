package testutils

import (
	"fmt"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertCounter(t *testing.T, metric *dto.Metric, value float64, desc string) {
	assert.Equal(t, value, metric.GetCounter().GetValue(), desc+": invalid Counter value")
}

func AssertGauge(t *testing.T, metric *dto.Metric, value float64, desc string) {
	assert.Equal(t, value, metric.GetGauge().GetValue(), desc+": invalid Gauge value")
}

func AssertHistogram(t *testing.T, metric *dto.Metric, sampleCount uint64, desc string) {
	assert.Equal(t, sampleCount, metric.GetHistogram().GetSampleCount(), desc+": invalid Histogram sample count")
}

func AssertMetricFamily(t *testing.T, family *dto.MetricFamily, name string, typ dto.MetricType, desc string, labels map[string]string) {
	assert.Equal(t, typ, family.GetType(), desc+": invalid metric type")
	assert.Equal(t, name, family.GetName(), desc+": invalid name")
}

func AssertCounterFamily(t *testing.T, family *dto.MetricFamily, namespace, name string, values []float64, desc string, labels map[string]string) {
	AssertMetricFamily(t, family, generateFamilyName(namespace, name), dto.MetricType_COUNTER, desc, labels)
	metrics := filterMetricFamily(family, labels)
	require.Len(t, metrics, len(values), desc+": invalid count of metrics")
	for i, metric := range metrics {
		AssertCounter(t, metric, values[i], desc+fmt.Sprintf(" (#%v)", i))
	}
}

func AssertGaugeFamily(t *testing.T, family *dto.MetricFamily, namespace, name string, values []float64, desc string, labels map[string]string) {
	AssertMetricFamily(t, family, fmt.Sprintf("%s_%s", namespace, name), dto.MetricType_GAUGE, desc, labels)
	metrics := filterMetricFamily(family, labels)
	require.Len(t, metrics, len(values), desc+": invalid count of metrics")
	for i, metric := range metrics {
		AssertGauge(t, metric, values[i], desc+fmt.Sprintf(" (#%v)", i))
	}
}

func AssertHistogramFamily(t *testing.T, family *dto.MetricFamily, namespace, name string, sampleCounts []uint64, desc string, labels map[string]string) {
	AssertMetricFamily(t, family, generateFamilyName(namespace, name), dto.MetricType_HISTOGRAM, desc, labels)
	metrics := filterMetricFamily(family, labels)
	require.Len(t, metrics, len(sampleCounts), desc+": invalid count of metrics")
	for i, metric := range metrics {
		AssertHistogram(t, metric, sampleCounts[i], desc+fmt.Sprintf(" (#%v)", i))
	}
}

func AssertFamilyValue(t *testing.T, families map[string]*dto.MetricFamily, namespace, name string, values interface{}, desc string, labels map[string]string) {
	mf, ok := families[generateFamilyName(namespace, name)]
	if !ok {
		assert.Error(t, fmt.Errorf("metric does not exists"))
		return
	}

	switch mf.GetType() {
	case dto.MetricType_COUNTER:
		AssertCounterFamily(t, mf, namespace, name, values.([]float64), desc, labels)
	case dto.MetricType_GAUGE:
		AssertGaugeFamily(t, mf, namespace, name, values.([]float64), desc, labels)
	case dto.MetricType_HISTOGRAM:
		AssertHistogramFamily(t, mf, namespace, name, values.([]uint64), desc, labels)
	default:
		assert.Error(t, fmt.Errorf("invalid metric type: %s", mf.GetType().String()))
	}
}
