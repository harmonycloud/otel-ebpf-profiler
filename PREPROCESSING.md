# Standalone preprocessing for Go 1.23

This fork is based on [OpenTelemetry eBPF Profiler](https://github.com/open-telemetry/opentelemetry-ebpf-profiler),
commit `d50432b35a5fabb8392d56d7d35e67fdd07f630d` (2026-10-12).
It makes the existing ELF, native/Go unwind and libc introspection packages usable
by Kindling without upgrading Kindling's OpenTelemetry SDK. Upstream attribution,
licenses and parsing algorithms are retained.

## Supported library surface

- `libpf/pfelf`: ELF access and address mapping.
- `nativeunwind/elfunwindinfo`: `.eh_frame`, `.debug_frame` and `.gopclntab` parsing.
- `libc`: glibc/musl TLS and thread-specific-data introspection.
- Their shared types, assembly analysis, remote-memory access and utility packages.

The default `go.mod` is the dependency manifest for this Go 1.23 library surface.
It has no OpenTelemetry SDK or cilium/ebpf dependency. Other packages in this
repository are retained for upstream synchronization; they are not covered by the
Go 1.23 library compatibility claim. In particular, Python is not adapted for
Kindling by this change, and this fork does not implement Kindling's BPF writer.
Callers must still validate generated unwind operations against their own BPF ABI.

## Changes from upstream

- `metrics` buffers and reports through a synchronous `func([]Metric)` callback.
  OTel instrument creation and export live in `metrics/otel`, used by the complete
  profiler. Counter addition and Gauge last-value behavior are preserved.
- BPF collection loading lives in `support/ebpfloader`. The shared `support`
  package only exposes types, constants and read-only embedded BPF data; importing
  preprocessing does not load or attach any BPF program.
- `Intern` owns input bytes before calling `unique.Make`. This protects Go 1.23.0
  hash keys from mutation when callers recycle a borrowed scan buffer. Existing
  strings still share canonical handles; calls incur a temporary string copy.
- Trace label hashing uses the Go 1.23 `maphash.Hash` API. Key length is included
  to separate key/value boundaries, and label ordering remains irrelevant.
- Newer benchmark/WaitGroup helpers are replaced in library tests. Generated
  fixtures can be built with the selected local toolchain and older binutils.
- `ExtractionOptions.DisableGoStackSwitch` lets consumers without a Go runtime
  unwinder use the existing PC/SP rules for `runtime.morestack` and stop at arm64
  `runtime.asmcgocall`. Default extraction still emits upstream's Go stack-switch
  commands. `ExtractELFWithOptions` and `NewStackDeltaProviderWithOptions` share
  all ELF parsing, validation and provider statistics with the default APIs.

Consumers can select that compatibility mode without changing the BPF map ABI:

```go
provider := elfunwindinfo.NewStackDeltaProviderWithOptions(
    elfunwindinfo.ExtractionOptions{DisableGoStackSwitch: true},
)
```

Rules still use the upstream representation. A consumer with a different wire
encoding must translate rules at its BPF writer boundary, including x86 return
address registers packed into the high nibble of `BaseReg`. Keep file/map lifetime
management and encoding conversion in the consumer rather than copying the parser.

## Use from another module

The canonical module path is retained. While testing a local checkout:

```sh
go mod edit -require=go.opentelemetry.io/ebpf-profiler@v0.0.0
go mod edit -replace=go.opentelemetry.io/ebpf-profiler=/absolute/path/to/this/fork
```

After publishing the fork, replace the local directory with a pinned fork version
or commit. Do not use an unpinned branch as the dependency contract. The default
manifest intentionally pins Go 1.23-compatible dependencies rather than inheriting
the complete profiler's SDK and toolchain requirements.

## Verify the library

Select an actual Go 1.23 toolchain in `PATH`; `GOTOOLCHAIN=local` prevents a silent
upgrade. GCC and objcopy are needed for test fixtures.

```sh
GOTOOLCHAIN=local make test-preprocess
GOTOOLCHAIN=local make test-preprocess CGO_ENABLED=1 GO_FLAGS=-race
(cd integrationtests/preprocess && GOTOOLCHAIN=local go test ./...)
```

The integration module deliberately uses OTel v1.8.0 / metric v0.25.0 alongside the
preprocessing library. It exercises real ELF/Go and libc parsing and verifies that
the old metrics API remains usable.

## Complete profiler and dependency maintenance

`profiler.mod` and `profiler.sum` retain the original complete-profiler dependency
set and Go 1.26 requirement. Use them when building/testing runtime packages:

```sh
go test -modfile=profiler.mod ./metrics/otel ./collector/... ./tracer/... ./processmanager/...
go mod tidy -modfile=profiler.mod
```

`make test` uses this complete-profiler manifest. Running an unqualified
`go test ./...` or `go mod tidy` at the repository root scans runtime-only packages
and would mix the two dependency sets; use the library targets above instead.
Update the default manifest only for the supported library package graph, then
run the old-SDK integration test. Keep the complete-profiler manifest separate.

When syncing upstream, review API/ABI changes and rerun both library compatibility
tests and runtime-adapter tests. Do not copy newer dependency requirements into the
library manifest merely because upstream changed them.
