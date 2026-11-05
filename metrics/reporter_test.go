// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"reflect"
	"testing"
)

func TestReporterPreservesBufferedMetrics(t *testing.T) {
	var received []Metric
	Start(func(batch []Metric) { received = append(received, batch...) })
	t.Cleanup(func() { Start(nil) })
	if !Enabled() {
		t.Fatal("reporting callback should enable collection")
	}
	want := []Metric{{ID: IDELFInfoCacheHit, Value: 3}, {ID: IDELFInfoCacheMiss, Value: 7}}
	mutex.Lock()
	nMetrics = 0
	clear(metricIDSet)
	mutex.Unlock()
	input := append([]Metric(nil), want...)
	input = append(input, Metric{ID: IDELFInfoCacheHit, Value: 99})
	AddSlice(input) // Keep the existing first-value-per-ID behavior in one timestamp.
	mutex.Lock()
	report()
	count := nMetrics
	mutex.Unlock()
	if !reflect.DeepEqual(received, want) || count != 0 {
		t.Fatalf("received=%v, buffered=%d; want %v and an empty buffer", received, count, want)
	}
	Start(nil)
	if Enabled() {
		t.Fatal("a nil callback must disable export")
	}
}
