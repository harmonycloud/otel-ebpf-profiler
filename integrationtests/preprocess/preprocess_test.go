// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package preprocess_test

import (
	"context"
	"os"
	"runtime"
	"testing"

	"go.opentelemetry.io/ebpf-profiler/libc"
	"go.opentelemetry.io/ebpf-profiler/libpf/pfelf"
	"go.opentelemetry.io/ebpf-profiler/nativeunwind/elfunwindinfo"
	legacymetric "go.opentelemetry.io/otel/metric"
)

func TestPreprocessingAlongsideLegacyMetrics(t *testing.T) {
	// This API was removed from modern OTel metrics. Compiling this call proves
	// importing preprocessing has not forced the consumer onto the new SDK.
	meter := legacymetric.NewNoopMeterProvider().Meter("legacy")
	counter, err := meter.NewInt64Counter("legacy_total")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(context.Background(), 1)
	if runtime.GOOS != "linux" {
		t.Skip("ELF integration requires Linux")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ref := pfelf.NewReference(executable, os.DirFS("/"))
	defer ref.Close()
	intervals, err := elfunwindinfo.ExtractELF(ref)
	if err != nil {
		t.Fatal(err)
	}
	if intervals.NumDeltas == 0 || len(intervals.Blocks) == 0 {
		t.Fatal("the real executable produced no unwind rules")
	}
	ef, err := ref.GetELF()
	if err != nil {
		t.Fatal(err)
	}
	goSymbols, err := elfunwindinfo.NewGopclntab(ef)
	if err != nil {
		t.Fatal(err)
	}
	defer goSymbols.Close()
	sym, err := goSymbols.LookupSymbol("main.main")
	if err != nil {
		t.Fatal(err)
	}
	file, line, name := goSymbols.Symbolize(uintptr(sym.Address))
	if file == "" || line == 0 || name != "main.main" {
		t.Fatalf("unexpected Go symbol: file=%q line=%d name=%q", file, line, name)
	}
}

func TestLibcPreprocessing(t *testing.T) {
	for _, name := range []string{
		"/lib/x86_64-linux-gnu/libc.so.6",
		"/lib/aarch64-linux-gnu/libc.so.6",
	} {
		if _, err := os.Stat(name); err != nil {
			continue
		}
		ef, err := pfelf.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		defer ef.Close()
		info, err := libc.ExtractLibcInfo(ef)
		if err != nil {
			t.Fatal(err)
		}
		if !info.HasTSDInfo() {
			t.Fatal("glibc preprocessing produced no thread-specific-data layout")
		}
		return
	}
	t.Skip("no glibc fixture at a known system path")
}
