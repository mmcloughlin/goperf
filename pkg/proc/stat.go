package proc

import (
	"os"

	"github.com/c9s/goprocinfo/linux"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

const selfstatfile = "/proc/self/stat"

// Stat is a config provider based on the `/proc/*/stat` file.
type Stat struct{}

// Key returns "procstat".
func (Stat) Key() cfg.Key { return "procstat" }

// Doc for the configuration provider.
func (Stat) Doc() string { return "Linux process status information from /proc/*/stat" }

// Available checks for the /proc/self/stat file.
func (Stat) Available() bool {
	_, err := os.Stat(selfstatfile)
	return err == nil
}

// Configuration reports performance-critical parameters from the /proc/self/stat file.
func (Stat) Configuration() (cfg.Configuration, error) {
	stat, err := linux.ReadProcessStat("/proc/self/stat")
	if err != nil {
		return nil, err
	}

	return cfg.Configuration{
		cfg.PerfProperty(
			"priority",
			"nice value from 0 (high) to 39 (low) or for real-time the negated scheduling priority minus one",
			cfg.IntValue(stat.Priority),
		),
		cfg.Property(
			"nice",
			"nice value in the range 19 (low priority) to -20 (high priority)",
			cfg.IntValue(stat.Nice),
		),
		cfg.Property(
			"rtpriority",
			"real-time priority in the range 1 to 99 or 0 for non-real-time",
			cfg.IntValue(stat.RtPriority),
		),
		cfg.PerfProperty(
			"policy",
			"scheduling policy",
			Policy(stat.Policy),
		),
	}, nil
}
