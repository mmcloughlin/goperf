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

	detectors := []Detector{
		// &Cohen{
		// 	WindowSize:    30,
		// 	MinEffectSize: 2.0,
		// },
		// &KZA{
		// 	M:                15,
		// 	K:                3,
		// 	PercentThreshold: 5,
		// },
		&Hybrid{
			WindowSize:    30,
			MinEffectSize: 2,

			M:                15,
			K:                3,
			PercentThreshold: 4,
			Context:          2,
		},
	}

	for _, filename := range filenames {
		filename := filename // scopelint
		for _, d := range detectors {
			d := d // scopelint
			name := filepath.Base(filename) + "/" + d.Name()
			t.Run(name, func(t *testing.T) {
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
}
