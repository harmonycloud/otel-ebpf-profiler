// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package libpf

import (
	"testing"

	"go.opentelemetry.io/ebpf-profiler/libpf/pfunsafe"
)

func TestInternOwnsBorrowedString(t *testing.T) {
	const original = "/intern-test/recycled-maps-buffer.so"
	buf := []byte(original)
	first := Intern(pfunsafe.ToString(buf))
	// /proc parsers recycle the backing bytes after their callback returns.
	clear(buf)
	if first.String() != original || Intern(original) != first {
		t.Fatal("interned value or canonical key changed after buffer reuse")
	}
}

func TestInternEmptyString(t *testing.T) {
	if Intern("") != NullString || NullString.String() != "" {
		t.Fatal("empty strings must retain the zero-value handle")
	}
}
