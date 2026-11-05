// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package support // import "go.opentelemetry.io/ebpf-profiler/support"

// TracerData returns the embedded BPF object for the current architecture.
// The returned bytes are shared and must not be modified. Collection loading
// lives in support/ebpfloader so preprocessing does not import cilium/ebpf.
func TracerData() []byte { return tracerData }
