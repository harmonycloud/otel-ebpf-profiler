// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build ignore

// Remove the ELF64 section table from a test fixture without requiring a recent
// GNU strip. PT_LOAD data, including Go build information, remains unchanged.
package main

import (
	"bytes"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: strip_sections input output")
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if len(data) < 64 || !bytes.Equal(data[:4], []byte{0x7f, 'E', 'L', 'F'}) || data[4] != 2 {
		log.Fatal("expected an ELF64 fixture")
	}
	clear(data[40:48]) // e_shoff
	clear(data[58:64]) // e_shentsize, e_shnum, e_shstrndx
	if err := os.WriteFile(os.Args[2], data, 0o755); err != nil {
		log.Fatal(err)
	}
}
