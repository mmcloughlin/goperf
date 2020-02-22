package sys

import (
	"fmt"
	"strconv"
	"time"

	cpuutil "github.com/shirou/gopsutil/cpu"
	hostutil "github.com/shirou/gopsutil/host"
	loadutil "github.com/shirou/gopsutil/load"
	memutil "github.com/shirou/gopsutil/mem"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

func init() {
	cfg.RegisterProvider(Host)
	cfg.RegisterProvider(VirtualMemory)
	cfg.RegisterProvider(CPU)
	cfg.RegisterProvider(LoadAverage)
}

var (
	Host          = cfg.NewProvider("host", "Host statistics.", host)
	VirtualMemory = cfg.NewProvider("mem", "Virtual memory usage statistics.", virtualmemory)
	CPU           = cfg.NewProvider("cpu", "CPU information", cpu)
	LoadAverage   = cfg.NewProvider("load", "Load averages", load)
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
		cfg.Property("num-procs", "number of processes", cfg.IntValue(info.Procs)),
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
		cfg.Property("total", "total amount of RAM on this system", cfg.BytesValue(vmem.Total)),
		cfg.Property("available", "RAM available for programs to allocate", cfg.BytesValue(vmem.Available)),
		cfg.Property("used", "RAM used by programs", cfg.BytesValue(vmem.Used)),
		cfg.Property("used-percent", "percentage of RAM used by programs", cfg.PercentageValue(vmem.UsedPercent)),
		cfg.Property("free", "kernel's measure of free memory", cfg.BytesValue(vmem.Free)),
	}, nil
}

func cpu() (cfg.Configuration, error) {
	procs, err := cpuutil.Info()
	if err != nil {
		return nil, err
	}

	c := cfg.Configuration{}
	for _, proc := range procs {
		c = append(c, processor(proc))
	}
	return c, nil
}

func processor(proc cpuutil.InfoStat) cfg.Entry {
	return cfg.Section(
		cfg.Key("cpu"+strconv.Itoa(int(proc.CPU))),
		fmt.Sprintf("processor %d information", proc.CPU),
		cfg.Property("vendorid", "vendor id", cfg.StringValue(proc.VendorID)),
		cfg.Property(
			"family",
			`identifies the type of processor in the system (for intel place the number in front of "86")`,
			cfg.StringValue(proc.Family),
		),
		cfg.Property("model", "model number", cfg.StringValue(proc.Model)),
		cfg.Property("stepping", "version number", cfg.IntValue(proc.Stepping)),
		cfg.Property("model-name", "common name of the processor", cfg.StringValue(proc.ModelName)),
		cfg.Property("physical-id", "physical processor number", cfg.StringValue(proc.PhysicalID)),
		cfg.Property("core-id", "physical core number within the processor", cfg.StringValue(proc.CoreID)),
		cfg.Property("cores", "number of physical cores", cfg.IntValue(proc.Cores)),
		cfg.Property("frequency", "nominal frequency", cfg.FrequencyValue(proc.Mhz*1e6)),
		cfg.Property("cache-size", "cache size (level 2)", cfg.BytesValue(proc.CacheSize*1e3)),
		cfg.Property("flags", "processor properties and feature sets", cfg.StringsValue(proc.Flags)),
	)
}
