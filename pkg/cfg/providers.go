package cfg

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	all atomic.Value
	mu  sync.Mutex
)

// RegisterProvider registers the configuration provider in the global store.
func RegisterProvider(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers, _ := all.Load().([]Provider)
	all.Store(append(providers, p))
}

// All returns all registered configuration providers.
func All() Providers {
	providers, _ := all.Load().([]Provider)
	return append(Providers{}, providers...)
}

// Providers is a list of providers.
type Providers []Provider

// Available returns true. Note that sub-providers will be checked for
// availability when Configuration() is called.
func (p Providers) Available() bool { return true }

// Keys returns all keys in the provider list.
func (p Providers) Keys() []string {
	keys := make([]string, len(p))
	for i := range p {
		keys[i] = string(p[i].Key())
	}
	return keys
}

// FilterAvailable returns the available sub-providers.
func (p Providers) FilterAvailable() Providers {
	a := make(Providers, 0, len(p))
	for i := range p {
		if p[i].Available() {
			a = append(a, p[i])
		}
	}
	return a
}

// Select returns the subset of providers with the given keys.
func (p Providers) Select(keys ...string) (Providers, error) {
	m := map[string]Provider{}
	for i := range p {
		m[string(p[i].Key())] = p[i]
	}

	s := make(Providers, len(keys))
	for i, k := range keys {
		if _, ok := m[k]; !ok {
			return nil, fmt.Errorf("provider %q not found", k)
		}
		s[i] = m[k]
	}
	return s, nil
}

// Configuration gathers configuration from all providers.
func (p Providers) Configuration() (Configuration, error) {
	c := make(Configuration, len(p))
	for i := range p {
		e, err := providerentry(p[i])
		if err != nil {
			return nil, err
		}
		c[i] = e
	}
	return c, nil
}

func providerentry(p Provider) (Entry, error) {
	section := SectionEntry{
		Labeled: p,
	}

	if !p.Available() {
		section.Sub = Configuration{
			Property(
				"available",
				fmt.Sprintf("availability of the %s provider", p.Key()),
				BoolValue(false),
			),
		}
	} else {
		sub, err := p.Configuration()
		if err != nil {
			return nil, err
		}
		section.Sub = sub
	}

	return section, nil
}
