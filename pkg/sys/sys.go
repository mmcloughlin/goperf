package sys

import (
	"time"

	hostutil "github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

var (
	Host          = cfg.NewProvider("host", "Host statistics.", host)
	VirtualMemory = cfg.NewProvider("mem", "Virtual memory usage statistics.", virtualmemory)
)

func host() (cfg.Configuration, error) {
	info, err := hostutil.Info()
	if err != nil {
		return nil, err
	}

	return cfg.Configuration{
		cfg.Property("hostname", "Hostname", cfg.StringValue(info.Hostname)),
		cfg.Property("uptime", "total uptime", time.Duration(info.Uptime)*time.Second),
		cfg.Property("boot-time", "boot timestamp", time.Unix(int64(info.BootTime), 0)),
		cfg.Property("num-procs", "number of processes", cfg.Uint64Value(info.Procs)),
		cfg.Property("os", "operating system", cfg.StringValue(info.OS)),
		cfg.Property("platform", "example: ubuntu", cfg.StringValue(info.Platform)),
		cfg.Property("platform-family", "example: debian", cfg.StringValue(info.PlatformFamily)),
		cfg.Property("platform-version", "version of the complete OS", cfg.StringValue(info.PlatformVersion)),
		cfg.Property("kernel-version", "version of the OS kernel", cfg.StringValue(info.KernelVersion)),
		cfg.Property("kernel-arch", "native cpu architecture queried at runtime", cfg.StringValue(info.KernelArch)),
		cfg.Property("virt-system", "virtualization system", cfg.StringValue(info.VirtualizationSystem)),
		cfg.Property("virt-role", "virtualization role", cfg.StringValue(info.VirtualizationRole)),
	}, nil
}

func virtualmemory() (cfg.Configuration, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return cfg.Configuration{
		cfg.Property(
			"total",
			"total amount of RAM on this system",
			cfg.BytesValue(vmem.Total),
		),
		cfg.Property(
			"available",
			"RAM available for programs to allocate",
			cfg.BytesValue(vmem.Available),
		),
		cfg.Property(
			"used",
			"RAM used by programs",
			cfg.BytesValue(vmem.Used),
		),
		cfg.Property(
			"used-percent",
			"percentage of RAM used by programs",
			cfg.PercentageValue(vmem.UsedPercent),
		),
		cfg.Property(
			"free",
			"kernel's measure of free memory",
			cfg.BytesValue(vmem.Free),
		),
	}, nil
}
