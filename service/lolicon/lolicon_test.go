package lolicon

import (
	"testing"

	"github.com/lz1998/lolicon_proxy/config"
)

func TestCallLolicon(t *testing.T) {
	resp, err := CallLolicon(config.Apikey, "1", "", "10")
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%+v", resp)
	}
}
