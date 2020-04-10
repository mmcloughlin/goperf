package coordinator

import "testing"

func TestValidateWorkerInvalid(t *testing.T) {
	invalid := []string{
		"",
		"123abc",
		"-----",
		"gopherPi",
	}
	for _, worker := range invalid {
		if err := validateWorker(worker); err == nil {
			t.Errorf("expected worker name %q to fail validation", worker)
		}
	}
}

func TestValidateWorkerValid(t *testing.T) {
	invalid := []string{
		"a",
		"abc-def-9",
		"gopher-raspberry-pi",
	}
	for _, worker := range invalid {
		if err := validateWorker(worker); err != nil {
			t.Errorf("worker name %q should be valid: got %v", worker, err)
		}
	}
}
