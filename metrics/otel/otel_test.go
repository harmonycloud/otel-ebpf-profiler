// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otel

import (
	"context"
	"testing"

	"go.opentelemetry.io/ebpf-profiler/metrics"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestReporterExportsCounterAndGauge(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	report := newReporter(provider.Meter("test"))
	var counter, gauge metrics.MetricDefinition
	for _, def := range metrics.GetDefinitions() {
		if def.Obsolete || def.Field == "" {
			continue
		}
		if def.Type == metrics.MetricTypeCounter && counter.ID == 0 {
			counter = def
		}
		if def.Type == metrics.MetricTypeGauge && gauge.ID == 0 {
			gauge = def
		}
	}
	report([]metrics.Metric{{ID: counter.ID, Value: 3}, {ID: gauge.ID, Value: 7}})
	report([]metrics.Metric{{ID: counter.ID, Value: 2}, {ID: gauge.ID, Value: 11}})
	var result metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &result); err != nil {
		t.Fatal(err)
	}
	var counterOK, gaugeOK bool
	for _, scope := range result.ScopeMetrics {
		for _, m := range scope.Metrics {
			switch m.Name {
			case counter.Field:
				data, ok := m.Data.(metricdata.Sum[int64])
				counterOK = ok && len(data.DataPoints) == 1 && data.DataPoints[0].Value == 5
			case gauge.Field:
				data, ok := m.Data.(metricdata.Gauge[int64])
				gaugeOK = ok && len(data.DataPoints) == 1 && data.DataPoints[0].Value == 11
			}
		}
	}
	if !counterOK || !gaugeOK {
		t.Fatalf("counter=%v gauge=%v; want cumulative sum 5 and last value 11", counterOK, gaugeOK)
	}
}

func TestNoopMeterDisablesReporter(t *testing.T) {
	if newReporter(noop.NewMeterProvider().Meter("test")) != nil {
		t.Fatal("a noop meter must not enable metric collection")
	}
}
