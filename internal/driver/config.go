package driver

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type ServiceConfig struct {
	TuyaConnectorInfo TuyaConnectorInfo
}

func (sw *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false //errors.New("unable to cast raw config to type 'ServiceConfig'")
	}

	*sw = *configuration

	return true
}

type ConnectionInfo struct {
	DeviceId string
}

func CreateConnectionInfo(protocols map[string]models.ProtocolProperties) (info *ConnectionInfo, err error) {
	info = new(ConnectionInfo)
	protocol, ok := protocols[Protocol]
	if !ok {
		return nil, fmt.Errorf("unable to load config, '%s' not exist", Protocol)
	}
	err = Load(protocol, info)

	return
}

type TuyaConnectorInfo struct {
	AccessId  string
	AccessKey string
	Region    string
}
