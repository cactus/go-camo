package maxprocs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/rdforte/gomaxecs/internal/config"
	ecstask "github.com/rdforte/gomaxecs/internal/task"
)

const maxProcsKey = "GOMAXPROCS"

var errPanic = errors.New("maxprocs: panic occurred while setting GOMAXPROCS")

// Set sets GOMAXPROCS based on the CPU limit of the container and the task.
// returns a function to reset GOMAXPROCS to its previous value and an error if one occurred.
// If the GOMAXPROCS environment variable is set, it will honor that value.
func Set(opts ...config.Option) (undo func(), err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", errPanic, r)
		}

		if undo == nil {
			undo = func() {}
		}
	}()

	cfg := config.New(opts...)
	task := ecstask.New(cfg)

	undo = func() {
		cfg.Logf("maxprocs: No GOMAXPROCS change to reset")
	}

	if procs, ok := shouldHonorGOMAXPROCSEnv(); ok {
		cfg.Logf("maxprocs: Honoring GOMAXPROCS=%q as set in environment", procs)
		return undo, nil
	}

	prevProcs := prevMaxProcs()
	undo = func() {
		cfg.Logf("maxprocs: Resetting GOMAXPROCS to %v", prevProcs)
		setMaxProcs(prevProcs)
	}

	procs, err := task.GetMaxProcs(context.Background())
	if err != nil {
		cfg.Logf("maxprocs: Failed to set GOMAXPROCS: %v", err)
		return undo, fmt.Errorf("failed to set GOMAXPROCS: %w", err)
	}

	cfg.DebugLogf("Calculated GOMAXPROCS to be %d", procs)

	setMaxProcs(procs)
	cfg.Logf("maxprocs: Updated GOMAXPROCS=%v", procs)

	return undo, nil
}

// shouldHonorGOMAXPROCSEnv returns the GOMAXPROCS environment variable if present
// and a boolean indicating if it should be honored.
func shouldHonorGOMAXPROCSEnv() (string, bool) {
	return os.LookupEnv(maxProcsKey)
}

func prevMaxProcs() int {
	return runtime.GOMAXPROCS(0)
}

func setMaxProcs(procs int) {
	runtime.GOMAXPROCS(procs)
}

// WithLogger sets the logger. By default, no logger is set.
func WithLogger(printf func(format string, args ...any)) config.Option {
	return config.WithLogger(printf)
}

// IsECS returns true if detected ECS environment.
func IsECS() bool {
	return len(config.GetECSMetadataURI()) > 0
}
