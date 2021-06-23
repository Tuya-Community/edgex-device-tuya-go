package driver

import (
	"github.com/tuya/tuya-connector-go/connector"
	"github.com/tuya/tuya-connector-go/connector/env"
	"github.com/tuya/tuya-connector-go/connector/httplib"
	"log"
	"os"
	"testing"
)

var deviceId string

func init() {
	accessId := os.Getenv("TUYA_TEST_ACCESS_ID")
	log.Println("TUYA_TEST_ACCESS_ID:", accessId)
	accessKey := os.Getenv("TUYA_TEST_ACCESS_KEY")
	log.Println("TUYA_TEST_ACCESS_KEY:", accessKey)
	region := os.Getenv("TUYA_TEST_REGION")
	log.Println("TUYA_TEST_REGION:", region)
	deviceId = os.Getenv("TUYA_TEST_DEVICE_ID")
	log.Println("TUYA_TEST_DEVICE_ID:", deviceId)
	if len(accessId) * len(accessKey) * len(region) * len(deviceId) == 0 {
		log.Fatal("Get empty env")
	}
	apiHost := ""
	msgHost := ""
	switch region {
	case RegionCN:
		apiHost = httplib.URL_CN
		msgHost = httplib.MSG_CN
	case RegionUS:
		apiHost = httplib.URL_US
		msgHost = httplib.MSG_US
	case RegionEU:
		apiHost = httplib.URL_EU
		msgHost = httplib.MSG_EU
	case RegionIN:
		apiHost = httplib.URL_IN
		msgHost = httplib.MSG_IN
	}
	connector.InitWithOptions(
		env.WithAccessID(accessId),
		env.WithAccessKey(accessKey),
		env.WithApiHost(apiHost),
		env.WithMsgHost(msgHost),
	)
}

func TestTuyaClient_GetValue(t *testing.T) {
	client, _ := NewClient(&ConnectionInfo{DeviceId: deviceId})

	v, err := client.GetValue(&GetParam{BaseParam{
		Code: "switch_1",
	}})

	if err != nil {
		t.Fatal(err)
	}

	t.Log(v)
}

func TestTuyaClient_SetValue(t *testing.T) {
	client, _ := NewClient(&ConnectionInfo{DeviceId: deviceId})
	err := client.SetValue(&SetParam{
		BaseParam: BaseParam{
			Code: "countdown_1",
		},
		Value: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("SUCCESS")
}
