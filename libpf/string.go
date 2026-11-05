// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package libpf // import "go.opentelemetry.io/ebpf-profiler/libpf"

import (
	"strings"
	"unique"
)

// String is an interned string. This is a wrapper for unique.Handle[string],
// but provides String() to be usable as printf, and also treats the default
// initializer as the empty string.
type String struct {
	value unique.Handle[string]
}

var NullString = String{unique.Handle[string]{}}

func Intern(str string) String {
	if str == "" {
		return NullString
	}
	// Callers may pass a string backed by a reusable /proc scan buffer.
	// Go 1.23.0 clones the stored value but retains the input as the map key.
	// Own the bytes before interning so subsequent buffer reuse cannot mutate it.
	return String{unique.Make(strings.Clone(str))}
}

func (s String) String() string {
	if s == NullString {
		return ""
	}
	return s.value.Value()
}
