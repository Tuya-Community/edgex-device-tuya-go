package driver

import (
	"github.com/tuya/tuya-connector-go/connector"
	"github.com/tuya/tuya-connector-go/connector/env"
	"testing"
)

func init() {
	connector.InitWithOptions(
		env.WithAccessID("cjshh8yqaoe7wdlsbl7q"),
		env.WithAccessKey("e51937d0b2644364b7d92d0a9bfe5e5b"),
		env.WithApiHost("https://openapi.tuyacn.com"),
		env.WithMsgHost("pulsar+ssl://mqe.tuyacn.com:7285/"),
	)
}

func TestTuyaClient_GetValue(t *testing.T) {
	client, _ := NewClient(&ConnectionInfo{DeviceId: "06870016bcddc237998d"})

	v, err := client.GetValue(&GetParam{BaseParam{
		Code: "switch_1",
	}})

	if err != nil {
		t.Fatal(err)
	}

	t.Log(v)
}

func TestTuyaClient_SetValue(t *testing.T) {
	client, _ := NewClient(&ConnectionInfo{DeviceId: "06870016bcddc237998d"})
	err := client.SetValue(&SetParam{
		BaseParam: BaseParam{
			Code: "switch_1",
		},
		Value: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("SUCCESS")
}
