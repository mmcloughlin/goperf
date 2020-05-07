// Package shield provides CPU isolation for benchmark execution.
package shield

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/mmcloughlin/cb/internal/errutil"
	"github.com/mmcloughlin/cb/pkg/cpuset"
)

// Shield uses cpusets to setup exclusive access to some CPUs.
type Shield struct {
	root    string      // root cpuset
	shield  string      // shield cpuset (relative to root)
	shieldn int         // number of cpus in shield cpuset (0 for max)
	sys     string      // system cpuset name (relative to root)
	sysn    int         // minimum number of cpus in system cpuset
	log     *zap.Logger // logger

	deferred []func() error
}

// Option configures a shield.
type Option func(*Shield)

// NewShield builds a CPU shield.
func NewShield(opts ...Option) *Shield {
	s := &Shield{
		root:   "",
		shield: "shield",
		sys:    "sys",
		sysn:   1,
		log:    zap.NewNop(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithRoot configures the root cpuset.
func WithRoot(name string) Option {
	return func(s *Shield) { s.root = name }
}

// WithShieldName configures the name of the shield cpuset. Note this is interpreted relative to the root.
func WithShieldName(name string) Option {
	return func(s *Shield) { s.shield = name }
}

// WithShieldNumCPU configures the number of CPUs assigned to the shield cpuset.
func WithShieldNumCPU(n int) Option {
	return func(s *Shield) { s.shieldn = n }
}

// WithSystemName configures the name of the system cpuset. Note this is interpreted
// relative to the root.
func WithSystemName(name string) Option {
	return func(s *Shield) { s.sys = name }
}

// WithSystemNumCPU configures the minimum number of CPUs assigned to the system cpuset.
func WithSystemNumCPU(n int) Option {
	return func(s *Shield) { s.sysn = n }
}

// WithLogger configures the logger for CPU shield operations.
func WithLogger(l *zap.Logger) Option {
	return func(s *Shield) { s.log = l.Named("shield") }
}

// Name of tuning method.
func (s *Shield) Name() string { return "shield" }

// ShieldName returns the name of the shield cpuset.
func (s *Shield) ShieldName() string {
	return s.shield
}

// Available reports whether the shield mechanism can be applied. Note this is a
// rudimentary check that the environment supports cpusets at all, it is still
// possible that applying the shield would error.
func (s *Shield) Available() bool {
	root := cpuset.NewCPUSet(s.root)
	allcpu, err := root.CPUs()
	return err == nil && len(allcpu) >= s.mincpus()
}

// Apply the configured shielding.
func (s *Shield) Apply() error {
	err := s.apply()
	// On error, attempt to cleanup.
	if err != nil {
		s.log.Debug("attempting reset")
		if err := s.Reset(); err != nil {
			s.log.Error("reset failed", zap.Error(err))
		}
	}
	return err
}

func (s *Shield) apply() error {
	// Determine available CPUs.
	root := cpuset.NewCPUSet(s.root)
	allcpu, err := root.CPUs()
	if err != nil {
		return err
	}
	s.log.Debug("fetched available cpus", zap.String("root", root.Path()), zap.Stringer("cpus", allcpu))

	if len(allcpu) <= s.sysn {
		return fmt.Errorf("not enough cpus: require %d for system but root has %d", s.sysn, len(allcpu))
	}

	// Assign CPUs to the two sets.
	shieldcpu, syscpu, err := s.assign(allcpu)
	if err != nil {
		return fmt.Errorf("could not assign cpus: %w", err)
	}

	s.log.Debug("chosen cpus",
		zap.Stringer("shield", shieldcpu),
		zap.Stringer("sys", syscpu),
	)

	// Create system cpuset.
	sys, err := cpuset.Create(s.sys)
	if err != nil {
		return err
	}
	s.cleanup(sys.Remove)

	// Assign CPUs.
	if err := sys.SetCPUs(syscpu); err != nil {
		return err
	}

	// Assign memory nodes.
	mems, err := root.Mems()
	if err != nil {
		return err
	}

	if err := sys.SetMems(mems); err != nil {
		return err
	}

	// Move all tasks from root to system.
	if err := s.movetasks(root, sys); err != nil {
		return err
	}
	s.cleanup(func() error {
		return s.movetasks(sys, root)
	})

	// Create shield cpuset.
	shield, err := cpuset.Create(s.shield)
	if err != nil {
		return err
	}
	s.cleanup(shield.Remove)

	// Configure shield CPUs.
	if err := shield.SetCPUs(shieldcpu); err != nil {
		return err
	}

	// Memory nodes.
	if err := shield.SetMems(mems); err != nil {
		return err
	}

	// Exclusive.
	if err := shield.EnableCPUExclusive(); err != nil {
		return err
	}

	return nil
}

// cleanup adds an operation to be called on Reset(). Cleanup functions will be
// called in reverse order, similar to defer.
func (s *Shield) cleanup(f func() error) {
	s.deferred = append(s.deferred, f)
}

// Reset restores cpuset configuration to the state prior to shielding.
func (s *Shield) Reset() error {
	var errs errutil.Errors
	for i := len(s.deferred) - 1; i >= 0; i-- {
		if err := s.deferred[i](); err != nil {
			errs.Add(err)
		}
	}
	return errs.Err()
}

// movetasks moves all tasks form src to dst, with additional logging.
func (s *Shield) movetasks(src, dst *cpuset.CPUSet) error {
	s.log.Debug("moving tasks", zap.String("src", src.Path()), zap.String("dst", dst.Path()))
	m, err := cpuset.MoveTasks(src, dst)
	if err != nil {
		return err
	}
	s.log.Debug("tasks moved",
		zap.Int("num_moved", len(m.Moved)),
		zap.Int("num_nonexistent", len(m.Nonexistent)),
		zap.Int("num_invalid", len(m.Invalid)),
	)
	return nil
}

// assign cpus to shield and system.
func (s *Shield) assign(cpus cpuset.Set) (shield, sys cpuset.Set, err error) {
	if len(cpus) < s.mincpus() {
		return nil, nil, errors.New("not enough cpus")
	}

	m := cpus.SortedMembers()

	// Shield CPUs 0 value means assign the max.
	n := s.shieldn
	if n == 0 {
		n = len(m) - s.sysn
	}

	return cpuset.NewSet(m[:n]...), cpuset.NewSet(m[n:]...), nil
}

// mincpus returns the minimum number of CPUs required to satisfy configuration.
func (s *Shield) mincpus() int {
	return max(s.shieldn, 1) + s.sysn
}

// max is integer maximum.
func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
