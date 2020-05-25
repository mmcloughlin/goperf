// Package sys provides system configuration and tuning.
package sys

import (
	"time"

	hostutil "github.com/shirou/gopsutil/host"
	loadutil "github.com/shirou/gopsutil/load"
	memutil "github.com/shirou/gopsutil/mem"

	"github.com/mmcloughlin/goperf/pkg/cfg"
)

var (
	Host          = cfg.NewProviderFunc("host", "Host statistics.", host)
	VirtualMemory = cfg.NewProviderFunc("mem", "Virtual memory usage statistics.", virtualmemory)
	CPU           = cfg.NewProviderFunc("cpu", "CPU information", cpu)
	LoadAverage   = cfg.NewProviderFunc("load", "Load averages", load)
)

func host() (cfg.Configuration, error) {
	info, err := hostutil.Info()
	if err != nil {
		return nil, err
	}

	return cfg.Configuration{
		cfg.Property("hostname", "Hostname", cfg.StringValue(info.Hostname)),
		cfg.Property("uptime", "total uptime", time.Duration(info.Uptime)*time.Second),
		cfg.Property("boottime", "boot timestamp", cfg.TimeValue(time.Unix(int64(info.BootTime), 0))),
		cfg.Property("numprocs", "number of processes", cfg.IntValue(info.Procs)),
		cfg.PerfProperty("os", "operating system", cfg.StringValue(info.OS)),
		cfg.Property("platform", "example: ubuntu", cfg.StringValue(info.Platform)),
		cfg.Property("platformfamily", "example: debian", cfg.StringValue(info.PlatformFamily)),
		cfg.Property("platformversion", "version of the complete OS", cfg.StringValue(info.PlatformVersion)),
		cfg.Property("kernelversion", "version of the OS kernel", cfg.StringValue(info.KernelVersion)),
		cfg.PerfProperty("kernelarch", "native cpu architecture queried at runtime", cfg.StringValue(info.KernelArch)),
		cfg.Property("virtsystem", "virtualization system", cfg.StringValue(info.VirtualizationSystem)),
		cfg.PerfProperty("virtrole", "virtualization role", cfg.StringValue(info.VirtualizationRole)),
	}, nil
}

func load() (cfg.Configuration, error) {
	avg, err := loadutil.Avg()
	if err != nil {
		return nil, err
	}

	return cfg.Configuration{
		cfg.Property("avg1", "1-minute load average", cfg.Float64Value(avg.Load1)),
		cfg.Property("avg5", "5-minute load average", cfg.Float64Value(avg.Load5)),
		cfg.Property("avg15", "15-minute load average", cfg.Float64Value(avg.Load15)),
	}, nil
}

func virtualmemory() (cfg.Configuration, error) {
	vmem, err := memutil.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return cfg.Configuration{
		cfg.PerfProperty("total", "total amount of RAM on this system", cfg.BytesValue(vmem.Total)),
		cfg.Property("available", "RAM available for programs to allocate", cfg.BytesValue(vmem.Available)),
		cfg.Property("used", "RAM used by programs", cfg.BytesValue(vmem.Used)),
		cfg.Property("usedpercent", "percentage of RAM used by programs", cfg.PercentageValue(vmem.UsedPercent)),
		cfg.Property("free", "kernel's measure of free memory", cfg.BytesValue(vmem.Free)),
	}, nil
}
