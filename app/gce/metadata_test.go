package gce

import (
	"testing"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

func TestMetadataImplementsProvider(t *testing.T) {
	var _ cfg.Provider = new(Metadata)
}
