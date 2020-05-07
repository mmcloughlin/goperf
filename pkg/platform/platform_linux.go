package platform

import (
	"flag"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/cb/pkg/command"
	"github.com/mmcloughlin/cb/pkg/runner"
	"github.com/mmcloughlin/cb/pkg/shield"
	"github.com/mmcloughlin/cb/pkg/sys"
	"github.com/mmcloughlin/cb/pkg/wrap"
)

type Platform struct {
	shieldname string
	shieldn    int
	sysname    string
	sysn       int

	base   command.Base
	cfg    subcommands.Command
	pri    subcommands.Command
	cpuset subcommands.Command
}

func New(b command.Base) *Platform {
	return &Platform{base: b}
}

func (p *Platform) Wrappers() []subcommands.Command {
	p.cfg = wrap.NewConfigDefault(p.base)
	p.pri = wrap.NewPrioritize(p.base)
	p.cpuset = wrap.NewCPUSet(p.base)
	return []subcommands.Command{p.cfg, p.pri, p.cpuset}
}

func (p *Platform) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.shieldname, "shield", "shield", "shield cpuset name")
	f.IntVar(&p.shieldn, "shieldnumcpu", 0, "number of cpus in shield cpuset (0 for max)")
	f.StringVar(&p.sysname, "sys", "sys", "system cpuset name")
	f.IntVar(&p.sysn, "sysnumcpu", 1, "minimum number of cpus in system cpuset")
}

// ConfigureRunner sets benchmark runner options.
func (p *Platform) ConfigureRunner(r *runner.Runner) error {
	// Apply static wrappers.
	for _, wrapper := range []subcommands.Command{p.cfg, p.pri} {
		w, err := wrap.RunUnder(wrapper)
		if err != nil {
			return err
		}
		r.Wrap(w)
	}

	// Apply tuning methods. Note SMT deactivation needs to come early since it
	// changes the number of CPUs on the platform.
	r.Tune(sys.DeactivateSMT{})
	r.Tune(sys.DisableIntelTurbo{})
	r.Tune(sys.SetScalingGovernor{Governor: "performance"})

	// Setup CPU shield.
	s := shield.NewShield(
		shield.WithShieldName(p.shieldname),
		shield.WithShieldNumCPU(p.shieldn),
		shield.WithSystemName(p.sysname),
		shield.WithSystemNumCPU(p.sysn),
		shield.WithLogger(p.base.Log),
	)
	r.Tune(s)

	w, err := wrap.RunUnderCPUSet(p.cpuset, p.shieldname)
	if err != nil {
		return err
	}
	r.Wrap(w)

	return nil
}
