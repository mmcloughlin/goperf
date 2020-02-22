package sys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

// Reference: https://github.com/torvalds/linux/blob/34dabd81160f7bfb18b67c1161b3c4d7ca6cab83/Documentation/ABI/testing/sysfs-devices-system-cpu#L339-L384
//
//	What:		/sys/devices/system/cpu/cpu*/cache/index*/<set_of_attributes_mentioned_below>
//	Date:		July 2014(documented, existed before August 2008)
//	Contact:	Sudeep Holla <sudeep.holla@arm.com>
//			Linux kernel mailing list <linux-kernel@vger.kernel.org>
//	Description:	Parameters for the CPU cache attributes
//
//			allocation_policy:
//				- WriteAllocate: allocate a memory location to a cache line
//						 on a cache miss because of a write
//				- ReadAllocate: allocate a memory location to a cache line
//						on a cache miss because of a read
//				- ReadWriteAllocate: both writeallocate and readallocate
//
//			attributes: LEGACY used only on IA64 and is same as write_policy
//
//			coherency_line_size: the minimum amount of data in bytes that gets
//					     transferred from memory to cache
//
//			level: the cache hierarchy in the multi-level cache configuration
//
//			number_of_sets: total number of sets in the cache, a set is a
//					collection of cache lines with the same cache index
//
//			physical_line_partition: number of physical cache line per cache tag
//
//			shared_cpu_list: the list of logical cpus sharing the cache
//
//			shared_cpu_map: logical cpu mask containing the list of cpus sharing
//					the cache
//
//			size: the total cache size in kB
//
//			type:
//				- Instruction: cache that only holds instructions
//				- Data: cache that only caches data
//				- Unified: cache that holds both data and instructions
//
//			ways_of_associativity: degree of freedom in placing a particular block
//						of memory in the cache
//
//			write_policy:
//				- WriteThrough: data is written to both the cache line
//						and to the block in the lower-level memory
//				- WriteBack: data is written only to the cache line and
//					     the modified cache line is written to main
//					     memory only when it is replaced
//

// Caches provides configuration about processor caches.
type Caches struct{}

// Key returns "cache".
func (Caches) Key() cfg.Key { return "cache" }

// Doc for the configuration provider.
func (Caches) Doc() string { return "Processor cache hierarchy" }

// Available checks whether the cache sysfs files are present.
func (Caches) Available() bool {
	info, err := os.Stat("/sys/devices/system/cpu/cpu0/cache/index0")
	return err == nil && info.IsDir()
}

// Configuration queries sysfs for Intel P-state configuration.
func (Caches) Configuration() (cfg.Configuration, error) {
	properties := []fileproperty{
		{"type", parsestring, "cache type: Data, Instruction or Unified"},
		{"level", parseint, "the cache hierarchy in the multi-level cache configuration"},
		{"size", parsesize, "total cache size"},
		{"coherency_line_size", parseint, "minimum amount of data in bytes that gets transferred from memory to cache"},
		{"ways_of_associativity", parseint, "degree of freedom in placing a particular block of memory in the cache"},
		{"number_of_sets", parseint, "total number of sets in the cache, a set is a collection of cache lines with the same cache index"},
		{"shared_cpu_list", parsestring, "list of logical cpus sharing the cache"}, // TODO(mbm): parse as string?
		{"physical_line_partition", parseint, "number of physical cache line per cache tag"},
	}

	cpudirs, err := filepath.Glob("/sys/devices/system/cpu/cpu*")
	if err != nil {
		return nil, err
	}

	c := cfg.Configuration{}
	for _, cpudir := range cpudirs {
		cpu := filepath.Base(cpudir)
		pattern := filepath.Join(cpudir, "cache", "index*")
		cachedirs, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		cachecfg := cfg.Configuration{}
		for _, cachedir := range cachedirs {
			cache := filepath.Base(cachedir)
			sub, err := parsefiles(cachedir, properties)
			if err != nil {
				return nil, err
			}
			section := cfg.Section(
				cfg.Key(cache),
				fmt.Sprintf("%s cache for %s", cache, cpu),
				sub...,
			)
			cachecfg = append(cachecfg, section)
		}

		c = append(c, cfg.Section(
			cfg.Key(cpu),
			fmt.Sprintf("caches for %s", cpu),
			cachecfg...,
		))
	}

	return c, nil
}

func parsesize(s string) (cfg.Value, error) {
	if len(s) == 0 || s[len(s)-1] != 'K' {
		return nil, errors.New("expected last character of size to be K")
	}
	b, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return nil, err
	}
	return cfg.BytesValue(b * 1024), nil
}
