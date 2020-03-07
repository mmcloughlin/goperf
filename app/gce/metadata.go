package gce

import (
	"net/http"
	"strings"

	"cloud.google.com/go/compute/metadata"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

// Metadata provides configuration based on Google Compute Engine metadata.
type Metadata struct {
	c *metadata.Client
}

// NewMetadata builds a GCE metadata configuration provider backed by the given HTTP client.
func NewMetadata(c *http.Client) *Metadata {
	return &Metadata{
		c: metadata.NewClient(c),
	}
}

// Key returns "gce".
func (Metadata) Key() cfg.Key { return "gce" }

// Doc describes this configuration providers.
func (Metadata) Doc() string { return "google compute engine metadata" }

// Available checks whether the processes is running on Google Compute Engine.
func (m *Metadata) Available() bool { return metadata.OnGCE() }

// Configuration queries metadata for Google Compute Engine configuration.
func (m *Metadata) Configuration() (cfg.Configuration, error) {
	properties := []struct {
		Key cfg.Key
		Doc string
		Get func() (string, error)
	}{
		{"name", "instance name", m.getter("instance/name")},
		{"zone", "google compute engine vm zone", m.c.Zone},
		{"cpuplatform", "google compute engine cpu platform", m.getter("instance/cpu-platform")},
		{"machinetype", "google compute engine machine type", m.machinetype},
		{"preemptible", "if the instance is preemptable", m.getter("instance/scheduling/preemptible")},
	}
	c := cfg.Configuration{}
	for _, p := range properties {
		s, err := p.Get()
		if err != nil {
			return nil, err
		}
		c = append(c, cfg.Property(p.Key, p.Doc, cfg.StringValue(s)))
	}
	return c, nil
}

func (m *Metadata) machinetype() (string, error) {
	machtype, err := m.gettrimmed("instance/machine-type")
	if err != nil {
		return "", err
	}
	return machtype[strings.LastIndex(machtype, "/")+1:], nil
}

func (m *Metadata) getter(suffix string) func() (string, error) {
	return func() (string, error) {
		return m.gettrimmed(suffix)
	}
}

func (m *Metadata) gettrimmed(suffix string) (string, error) {
	s, err := m.c.Get(suffix)
	return strings.TrimSpace(s), err
}
