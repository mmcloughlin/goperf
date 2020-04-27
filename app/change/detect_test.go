package change

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/mmcloughlin/cb/app/trace"
)

type TestCase struct {
	Expect []int        `json:"expect"`
	Series trace.Series `json:"series"`
}

func TestDetectTestData(t *testing.T) {
	filenames, err := filepath.Glob("testdata/*.json")
	if err != nil {
		t.Fatal(err)
	}

	d := &Detector{
		WindowSize:    30,
		MinEffectSize: 2.0,
	}

	for _, filename := range filenames {
		filename := filename // scopelint
		t.Run(filepath.Base(filename), func(t *testing.T) {
			// Read JSON.
			b, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}

			var tc TestCase
			if err := json.Unmarshal(b, &tc); err != nil {
				t.Fatal(err)
			}

			// Detect changes.
			changes := d.Detect(tc.Series)

			t.Logf("expect: %v", tc.Expect)
			for _, c := range changes {
				t.Logf("%d: %v", c.CommitIndex, c.EffectSize)
			}
		})
	}
}
