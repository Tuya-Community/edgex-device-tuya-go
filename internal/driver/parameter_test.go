package driver

import (
	"testing"
)

func TestNewGetParam(t *testing.T) {
	attr := map[string]interface{}{
		Code: "1234567ui",
	}

	param, err := NewGetParam(attr)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(param)
}

func TestNewSetParam(t *testing.T) {
	attr := map[string]interface{}{
		Code: "1234567ui",
	}

	param, err := NewSetParam(attr, true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(param)
}
