package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/file-conversor/file_conversor/internal/cli"
)

const appName = "file_conversor"

func main() {
	// Profile the application if requested via environment variables
	stopProfiling := initProfiling()

	// Run the application and handle any errors
	exitCode, _ := cli.Run(appName)

	// stop profiling and exit with the appropriate code
	stopProfiling()
	os.Exit(exitCode)
}

// initProfiling checks environment variables to determine if profiling
// should be enabled. If profiling is enabled, it returns a function that
// should be deferred to stop the profiler and clean up resources.
func initProfiling() func() {
	// Check if profiling is enabled via environment variable
	if _, exists := os.LookupEnv("FC_PROFILE"); !exists {
		return func() {}
	}
	fmt.Fprintln(os.Stderr, "[profiling] starting ...")

	// cpu profile
	cpuPath := "cpu.prof"
	fCpuProf, cpuProfErr := os.Create(cpuPath)
	if cpuProfErr == nil {
		defer pprof.StartCPUProfile(fCpuProf)
		fmt.Fprintln(os.Stderr, "[profiling] CPU -> "+cpuPath)
	}

	// heap profile
	heapPath := "heap.prof"
	fmt.Fprintln(os.Stderr, "[profiling] heap -> "+heapPath)

	// trace profile
	tracePath := "trace.trace"
	fTraceProf, traceProfErr := os.Create(tracePath)
	if traceProfErr == nil {
		defer trace.Start(fTraceProf)
		fmt.Fprintln(os.Stderr, "[profiling] trace -> "+tracePath)
	}

	// return a function to stop profiling and clean up resources
	return func() {
		// cpu profile
		if cpuProfErr == nil {
			pprof.StopCPUProfile()
			fCpuProf.Close()
		}

		// heap profile
		if fHeapProf, err := os.Create(heapPath); err == nil {
			runtime.GC()
			pprof.WriteHeapProfile(fHeapProf)
			fHeapProf.Close()
		}

		// trace profile
		if traceProfErr == nil {
			trace.Stop()
			fTraceProf.Close()
		}

		// log profiling stopped
		fmt.Fprintln(os.Stderr, "[profiling] stopped")
	}
}
