package db

import (
	"testing"

	"github.com/mmcloughlin/cb/app/entity"
)

func TestTaskTypeMapping(t *testing.T) {
	for _, typ := range entity.TaskTypeValues() {
		typ := typ // scopelint
		t.Run(typ.String(), func(t *testing.T) {
			dbtyp, err := toTaskType(typ)
			if err != nil {
				t.Fatal(err)
			}
			roundtrip, err := mapTaskType(dbtyp)
			if err != nil {
				t.Fatal(err)
			}
			if roundtrip != typ {
				t.Fatal("roundtrip mismatch")
			}
		})
	}
}

func TestTaskStatusMapping(t *testing.T) {
	for _, status := range entity.TaskStatusValues() {
		status := status // scopelint
		t.Run(status.String(), func(t *testing.T) {
			dbstatus, err := toTaskStatus(status)
			if err != nil {
				t.Fatal(err)
			}
			roundtrip, err := mapTaskStatus(dbstatus)
			if err != nil {
				t.Fatal(err)
			}
			if roundtrip != status {
				t.Fatal("roundtrip mismatch")
			}
		})
	}
}
