// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package otel adapts profiler metrics to the OpenTelemetry metrics API.
// Preprocessing users can import metrics without importing this package.
package otel // import "go.opentelemetry.io/ebpf-profiler/metrics/otel"

import (
	"context"
	"fmt"

	"go.opentelemetry.io/ebpf-profiler/internal/log"
	"go.opentelemetry.io/ebpf-profiler/metrics"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// Start configures OTel export before metric-producing goroutines are started.
func Start(meter metric.Meter) {
	metrics.Start(newReporter(meter))
}

func newReporter(meter metric.Meter) func([]metrics.Metric) {
	if _, disabled := meter.(noop.Meter); disabled {
		return nil
	}
	counters := make(map[metrics.MetricID]metric.Int64Counter)
	gauges := make(map[metrics.MetricID]metric.Int64Gauge)
	for _, md := range metrics.GetDefinitions() {
		if md.Obsolete || md.Field == "" {
			continue
		}
		switch md.Type {
		case metrics.MetricTypeCounter:
			counter, err := meter.Int64Counter(md.Field,
				metric.WithDescription(md.Description), metric.WithUnit(md.Unit))
			if err != nil {
				log.Errorf("Creating Int64Counter: %v", err)
				continue
			}
			counters[md.ID] = counter
		case metrics.MetricTypeGauge:
			gauge, err := meter.Int64Gauge(md.Field,
				metric.WithDescription(md.Description), metric.WithUnit(md.Unit))
			if err != nil {
				log.Errorf("Creating Int64Gauge: %v", err)
				continue
			}
			gauges[md.ID] = gauge
		default:
			panic(fmt.Sprintf("Unknown metric type: %v", md.Type))
		}
	}
	return func(batch []metrics.Metric) {
		ctx := context.Background()
		for _, m := range batch {
			if counter, ok := counters[m.ID]; ok {
				counter.Add(ctx, int64(m.Value))
			}
			if gauge, ok := gauges[m.ID]; ok {
				gauge.Record(ctx, int64(m.Value))
			}
		}
	}
}
