// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package elfunwindinfo

import (
	"debug/elf"
	"os"
	"testing"

	"go.opentelemetry.io/ebpf-profiler/libpf/pfelf"
	sdtypes "go.opentelemetry.io/ebpf-profiler/nativeunwind/stackdeltatypes"
	"go.opentelemetry.io/ebpf-profiler/support"
	"go.opentelemetry.io/ebpf-profiler/testsupport"
)

func TestGoStackSwitchSelection(t *testing.T) {
	for _, arch := range []elf.Machine{elf.EM_X86_64, elf.EM_AARCH64} {
		for _, useFP := range []bool{false, true} {
			defaults := ExtractionOptions{}
			compatible := ExtractionOptions{DisableGoStackSwitch: true}
			morestack := getFunctionUnwindInfo("runtime.morestack", arch, useFP, defaults)
			if useFP && *morestack != sdtypes.UnwindInfoGoMorestack {
				t.Fatalf("default morestack changed on %s", arch)
			}
			if !useFP && *morestack != sdtypes.UnwindInfoStop {
				t.Fatalf("morestack without frame pointers must stop on %s", arch)
			}
			if getFunctionUnwindInfo("runtime.morestack", arch, useFP, compatible) != nil {
				t.Fatalf("morestack must fall back to PC/SP rules on %s", arch)
			}
			asmcgocall := getFunctionUnwindInfo("runtime.asmcgocall", arch, useFP, compatible)
			want := sdtypes.UnwindInfoStop
			if arch == elf.EM_X86_64 && useFP {
				want = sdtypes.UnwindInfoFramePointer
			}
			if *asmcgocall != want {
				t.Fatalf("compatible asmcgocall on %s: %+v", arch, *asmcgocall)
			}
		}
	}
}

func TestExtractionOptionsReachProvider(t *testing.T) {
	for _, file := range []string{"testdata/helloworld", "testdata/helloworld.arm64"} {
		testsupport.RequireGeneratedTestFile(t, file)
		// References name absolute paths; make the fixture visible in its root.
		ref := pfelf.NewReference("/"+file, os.DirFS("."))
		defaults, err := ExtractELF(ref)
		if err != nil {
			t.Fatal(err)
		}
		provider := NewStackDeltaProviderWithOptions(ExtractionOptions{DisableGoStackSwitch: true})
		compatible, err := provider.GetIntervalDataForFile(ref)
		ref.Close()
		if err != nil || compatible.NumDeltas == 0 {
			t.Fatalf("compatible extraction for %s: %v", file, err)
		}
		countSpecial := func(data *sdtypes.IntervalData) int {
			count := 0
			for _, block := range data.Blocks {
				for _, delta := range block.Deltas {
					if delta.Info.Flags&support.UnwindFlagCommand != 0 &&
						(delta.Info.Param == support.UnwindCommandGoMorestack ||
							delta.Info.Param == support.UnwindCommandGoAsmcgocall) {
						count++
					}
				}
			}
			return count
		}
		if countSpecial(defaults) == 0 || countSpecial(compatible) != 0 {
			t.Fatalf("command selection for %s: default=%d compatible=%d", file,
				countSpecial(defaults), countSpecial(compatible))
		}
		stats := provider.GetAndResetStatistics()
		if stats.Success != 1 || stats.ExtractionErrors != 0 || provider.GetAndResetStatistics().Success != 0 {
			t.Fatalf("provider statistics changed: %+v", stats)
		}
	}
}
