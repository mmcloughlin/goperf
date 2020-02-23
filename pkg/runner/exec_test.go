package runner

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestRunUnder(t *testing.T) {
	ru := RunUnder("/bin/wrap", "-flag", "arg")
	cases := []struct {
		Cmd        *exec.Cmd
		ExpectPath string
		ExpectArgs []string
	}{
		{
			Cmd: &exec.Cmd{
				Path: "/path/to/bin",
				Args: []string{"bin", "a", "b", "c"},
			},
			ExpectPath: "/bin/wrap",
			ExpectArgs: []string{
				"/bin/wrap", "-flag", "arg", "--", "/path/to/bin", "a", "b", "c",
			},
		},
		{
			Cmd: &exec.Cmd{
				Path: "/path/to/bin",
				Args: nil,
			},
			ExpectPath: "/bin/wrap",
			ExpectArgs: []string{
				"/bin/wrap", "-flag", "arg", "--", "/path/to/bin",
			},
		},
	}
	for _, c := range cases {
		ru(c.Cmd)
		if c.Cmd.Path != c.ExpectPath {
			t.Errorf("got path %q; expect %q", c.Cmd, c.ExpectPath)
		}
		if !reflect.DeepEqual(c.Cmd.Args, c.ExpectArgs) {
			t.Errorf("got args %q; expect %q", c.Cmd.Args, c.ExpectArgs)
		}
	}
}
