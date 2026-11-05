module example.com/preprocess-compat

go 1.23.0

require (
	go.opentelemetry.io/ebpf-profiler v0.0.0
	go.opentelemetry.io/otel v1.8.0
	go.opentelemetry.io/otel/metric v0.25.0
)

require (
	github.com/elastic/go-freelru v0.16.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	golang.org/x/arch v0.11.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
)

replace go.opentelemetry.io/ebpf-profiler => ../..
