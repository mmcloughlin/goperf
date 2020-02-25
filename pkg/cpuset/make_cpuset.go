// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
)

type Type struct {
	Name       string
	Go         string
	WriteVerb  string
	DefaultVar string
}

var (
	Flag = Type{"flag", "bool", "Set", "enabled"}
	Set  = Type{"list", "Set", "Set", "s"}
	Int  = Type{"int", "int", "Set", "n"}
	Ints = Type{"ints", "[]int", "Add", "ns"}
)

type Property struct {
	Filename     string
	FunctionName string
	Type         Type
	Var          string
	ReadOnly     bool
	Doc          []string
}

func (p Property) WriteFunc() string {
	return p.Type.WriteVerb + p.FunctionName
}

func (p Property) Variable() string {
	if p.Var != "" {
		return p.Var
	}
	return p.Type.DefaultVar
}

var Properties = []Property{
	{
		Filename:     "tasks",
		FunctionName: "Tasks",
		Type:         Ints,
		Var:          "tasks",
		Doc: []string{
			`Tasks returns the list of process IDs (PIDs) of the processes in the cpuset.`,
		},
	},
	{
		Filename:     "notify_on_release",
		FunctionName: "NotifyOnRelease",
		Type:         Flag,
		Doc: []string{
			`NotifyOnRelease reports whether the notify_on_release flag is set for this`,
			`cpuset. If true, that cpuset will receive special handling after it is`,
			`released, that is, after all processes cease using it (i.e., terminate or are`,
			`moved to a different cpuset) and all child cpuset directories have been`,
			`removed.`,
		},
	},
	{
		Filename:     "cpuset.cpus",
		FunctionName: "CPUs",
		Type:         Set,
		Doc: []string{
			`CPUs returns the set of physical numbers of the CPUs on which processes in`,
			`the cpuset are allowed to execute.`,
		},
	},
	{
		Filename:     "cpuset.cpu_exclusive",
		FunctionName: "CPUExclusive",
		Type:         Flag,
		Doc: []string{
			`CPUExclusive reports whether the cpuset has exclusive use of its CPUs (no`,
			`sibling or cousin cpuset may overlap CPUs). By default, this is off. Newly`,
			`created cpusets also initially default this to off.`,
			``,
			`Two cpusets are sibling cpusets if they share the same parent cpuset in the`,
			`hierarchy. Two cpusets are cousin cpusets if neither is the ancestor of the`,
			`other. Regardless of the cpu_exclusive setting, if one cpuset is the ancestor`,
			`of another, and if both of these cpusets have nonempty cpus, then their cpus`,
			`must overlap, because the cpus of any cpuset are always a subset of the cpus`,
			`of its parent cpuset.`,
		},
	},
	{
		Filename:     "cpuset.mems",
		FunctionName: "Mems",
		Type:         Set,
		Doc: []string{
			`Mems returns the list of memory nodes on which processes in this cpuset are`,
			`allowed to allocate memory.`,
		},
	},
	{
		Filename:     "cpuset.mem_exclusive",
		FunctionName: "MemExclusive",
		Type:         Flag,
		Doc: []string{
			`MemExclusive reports whether the cpuset has exclusive use of its memory nodes`,
			`(no sibling or cousin may overlap). Also if set, the cpuset is a Hardwall`,
			`cpuset. By default, this is off. Newly created cpusets also initially default`,
			`this to off.`,
			``,
			`Regardless of the mem_exclusive setting, if one cpuset is the ancestor of`,
			`another, then their memory nodes must overlap, because the memory nodes of`,
			`any cpuset are always a subset of the memory nodes of that cpuset's parent`,
			`cpuset.`,
		},
	},
	{
		Filename:     "cpuset.mem_hardwall",
		FunctionName: "MemHardwall",
		Type:         Flag,
		Doc: []string{
			`MemHardwall reports whether the cpuset is a Hardwall cpuset (see below).`,
			`Unlike mem_exclusive, there is no constraint on whether cpusets marked`,
			`mem_hardwall may have overlapping memory nodes with sibling or cousin`,
			`cpusets. By default, this is off. Newly created cpusets also initially`,
			`default this to off.`,
		},
	},
	{
		Filename:     "cpuset.memory_migrate",
		FunctionName: "MemoryMigrate",
		Type:         Flag,
		Doc: []string{
			`MemoryMigrate reports whether memory migration is enabled.`,
		},
	},
	{
		Filename:     "cpuset.memory_pressure",
		FunctionName: "MemoryPressure",
		Type:         Int,
		ReadOnly:     true,
		Doc: []string{
			`MemoryPressure reports a measure of how much memory pressure the processes in`,
			`this cpuset are causing. If MemoryPressureEnabled() is false this will always`,
			`be 0.`,
		},
	},
	{
		Filename:     "cpuset.memory_pressure_enabled",
		FunctionName: "MemoryPressureEnabled",
		Type:         Flag,
		Doc: []string{
			`MemoryPressureEnabled reports whether memory pressure calculations are`,
			`enabled for all cpusets in the system. This method only works for the root`,
			`cpuset. By default, this is off.`,
		},
	},
	{
		Filename:     "cpuset.memory_spread_page",
		FunctionName: "MemorySpreadPage",
		Type:         Flag,
		Doc: []string{
			`MemorySpreadPage reports whether pages in the kernel page cache`,
			`(filesystem buffers) are uniformly spread across the cpuset.`,
			`By default, this is off (0) in the top cpuset, and inherited`,
			`from the parent cpuset in newly created cpusets.`,
		},
	},
	{
		Filename:     "cpuset.memory_spread_slab",
		FunctionName: "MemorySpreadSlab",
		Type:         Flag,
		Doc: []string{
			`MemorySpreadSlab reports whether the kernel slab caches for file I/O`,
			`(directory and inode structures) are uniformly spread across the cpuset. By`,
			`default, this is off (0) in the top cpuset, and inherited from the parent`,
			`cpuset in newly created cpusets.`,
		},
	},
	{
		Filename:     "cpuset.sched_load_balance",
		FunctionName: "SchedLoadBalance",
		Type:         Flag,
		Doc: []string{
			`SchedLoadBalance reports wether the kernel will`,
			`automatically load balance processes in that cpuset over the`,
			`allowed CPUs in that cpuset.  If false the kernel will`,
			`avoid load balancing processes in this cpuset, unless some`,
			`other cpuset with overlapping CPUs has its sched_load_balance`,
			`flag set.`,
		},
	},
	{
		Filename:     "cpuset.sched_relax_domain_level",
		FunctionName: "SchedRelaxDomainLevel",
		Type:         Int,
		Var:          "level",
		Doc: []string{
			`SchedRelaxDomainLevel controls the width of the range of CPUs over which the`,
			`kernel scheduler performs immediate rebalancing of runnable tasks across`,
			`CPUs. If sched_load_balance is disabled, then the setting of`,
			`sched_relax_domain_level does not matter, as no such load balancing is done.`,
			`If sched_load_balance is enabled, then the higher the value of the`,
			`sched_relax_domain_level, the wider the range of CPUs over which immediate`,
			`load balancing is attempted.`,
		},
	},
}

func main() {
	if err := mainerr(); err != nil {
		log.Fatal(err)
	}
}

var (
	output = flag.String("output", "", "path to output file (default stdout)")
)

func mainerr() error {
	flag.Parse()

	// Generate.
	g := NewGenerator("cpuset", "CPUSet", "c")
	g.Methods(Properties)
	src, err := g.Format()
	if err != nil {
		return err
	}

	// Output.
	w := os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		w = f
	}

	_, err = w.Write(src)
	return err
}

type Generator struct {
	bytes.Buffer

	pkg      string
	name     string
	receiver string
}

func NewGenerator(pkg, name, receiver string) *Generator {
	return &Generator{
		pkg:      pkg,
		name:     name,
		receiver: receiver,
	}
}

func (g *Generator) Format() ([]byte, error) {
	return format.Source(g.Bytes())
}

func (g *Generator) Methods(ps []Property) {
	g.Printf("package %s\n", g.pkg)
	for _, p := range ps {
		g.Property(p)
	}
}

func (g *Generator) Property(p Property) {
	// Get method has the main documentation block.
	g.NL()
	for _, line := range p.Doc {
		g.Linef("// %s", line)
	}
	g.Linef("//")
	g.Linef("// Corresponds to the %q file in the cpuset directory.", p.Filename)

	g.Linef("func (%s *%s) %s() (%s, error) {", g.receiver, g.name, p.FunctionName, p.Type.Go)
	g.Linef("\treturn %sfile(%s.path(%q))", p.Type.Name, g.receiver, p.Filename)
	g.Linef("}")

	// Write method.
	if !p.ReadOnly {
		g.NL()
		g.Linef("// %s writes to the %q file of the cpuset.", p.WriteFunc(), p.Filename)
		g.Linef("//")
		g.Linef("// See %s() for the meaning of this field.", p.FunctionName)
		g.Linef("func (%s *%s) %s(%s %s) error {", g.receiver, g.name, p.WriteFunc(), p.Variable(), p.Type.Go)
		g.Linef("\treturn write%sfile(%s.path(%q), %s)", p.Type.Name, g.receiver, p.Filename, p.Variable())
		g.Linef("}")
	}
}

func (g *Generator) Printf(format string, a ...interface{}) {
	fmt.Fprintf(g, format, a...)
}

func (g *Generator) Linef(format string, a ...interface{}) {
	g.Printf(format+"\n", a...)
}

func (g *Generator) NL() { g.Linef("") }
