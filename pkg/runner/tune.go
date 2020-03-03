package runner

// Tuner tunes a system for benchmarking.
type Tuner interface {
	// Available checks whether the method can be applied at all. Note this is
	// intended to be a basic environment check, it is still possible that
	// Apply() could fail if Available() returns true.
	Available() bool

	// Apply the tuning method.
	Apply() error

	// Reset state to the same as before the last call to Apply().
	Reset() error
}
