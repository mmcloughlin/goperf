package mod

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mmcloughlin/cb/internal/test"
)

func TestModuleProxyNetworkInfo(t *testing.T) {
	test.RequiresNetwork(t)

	mdb := NewOfficialModuleProxy(http.DefaultClient)
	info, err := mdb.Info(context.Background(), "golang.org/x/crypto", "0f24fbd83dfbb33be4b41327d5a857464b89e3cd")
	if err != nil {
		t.Fatal(err)
	}

	expect := &RevInfo{
		Version: "v0.0.0-20200221170553-0f24fbd83dfb",
		Time:    time.Date(2020, 2, 21, 17, 05, 53, 0, time.UTC),
	}

	if info.Version != expect.Version {
		t.Error("version mismatch")
	}

	if !info.Time.Equal(expect.Time) {
		t.Error("time mismatch")
	}
}

func TestModuleProxyNetworkLatest(t *testing.T) {
	test.RequiresNetwork(t)

	mdb := NewOfficialModuleProxy(http.DefaultClient)
	info, err := mdb.Latest(context.Background(), "golang.org/x/crypto")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(info)
}
