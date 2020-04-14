package runner

// Tuner tunes a system for benchmarking.
type Tuner interface {
	// Name identifies the tuning method.
	Name() string

	// Available checks whether the method can be applied at all. Note this is
	// intended to be a basic environment check, it is still possible that
	// Apply() could fail if Available() returns true.
	Available() bool

	// Apply the tuning method.
	Apply() error

	// Reset state to default configuration.
	Reset() error
}
