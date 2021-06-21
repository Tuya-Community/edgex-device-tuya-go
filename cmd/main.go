package main

import (
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/startup"
	device_tuya_go "github.com/tuya/device-tuya-go"
	"github.com/tuya/device-tuya-go/internal/driver"
)

const (
	serviceName string = "device-tuya"
)

func main() {
	sd := driver.NewProtocolDriver()
	startup.Bootstrap(serviceName, device_tuya_go.Version, sd)
}
