module go.opentelemetry.io/ebpf-profiler

// Standalone preprocessing dependencies. Runtime packages use profiler.mod.
go 1.23.0

require (
	github.com/elastic/go-freelru v0.16.0
	github.com/google/uuid v1.6.0
	github.com/minio/sha256-simd v1.0.1
	github.com/stretchr/testify v1.10.0
	github.com/zeebo/xxh3 v1.1.0
	golang.org/x/arch v0.11.0
	golang.org/x/sys v0.31.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
