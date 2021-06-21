package driver

import (
	"context"
	"errors"
	"fmt"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/kainhuck/cache"
	"github.com/tuya/tuya-connector-go/connector"
	"github.com/tuya/tuya-connector-go/connector/env"
	"github.com/tuya/tuya-connector-go/connector/httplib"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var driver *Driver

type BaseDriver struct {
	Logger        logger.LoggingClient
	AsyncCh       chan<- *sdkModel.AsyncValues
	serviceConfig *ServiceConfig
}

type ClientMap struct {
	clients *cache.Cache
}

func (cm *ClientMap) Load(key string) (client Client, ok bool) {
	clientI, ok := cm.clients.Get(key)
	if ok {
		if clientI != nil {
			client = clientI.(Client)
		} else {
			return nil, false
		}
	}
	return
}

func (cm *ClientMap) Store(key string, value Client) {
	cm.clients.SetByDefaultExpiration(key, value)
	return
}

func (cm *ClientMap) Grow(key string, d time.Duration) {
	cm.clients.Grow(key, d)
}

type Driver struct {
	BaseDriver
	clientMap ClientMap
	cancel    context.CancelFunc
}

func NewProtocolDriver() sdkModel.ProtocolDriver {
	once.Do(func() {
		driver = new(Driver)
	})
	return driver
}

func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, deviceCh chan<- []sdkModel.DiscoveredDevice) error {
	d.Logger = lc
	d.AsyncCh = asyncCh
	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	ca := cache.New(ctx, 3*time.Minute, 5*time.Minute)
	ca.RegisterBeforeDelete(func(key string, value interface{}) {
		client := value.(Client)
		client.Close()
	})

	d.clientMap = ClientMap{clients: ca}

	d.serviceConfig = &ServiceConfig{}
	ds := service.RunningService()
	if err := ds.LoadCustomConfig(d.serviceConfig, CustomConfigSectionName); err != nil {
		return fmt.Errorf("unable to load '%s' custom configuration: %s", CustomConfigSectionName, err.Error())
	}

	//
	region := strings.ToUpper(d.serviceConfig.TuyaConnectorInfo.Region)
	apiHost := ""
	msgHost := ""
	switch region {
	case "CN":
		apiHost = httplib.URL_CN
		msgHost = httplib.MSG_CN
	case "US":
		apiHost = httplib.URL_US
		msgHost = httplib.MSG_US
	case "EU":
		apiHost = httplib.URL_EU
		msgHost = httplib.MSG_EU
	case "IN":
		apiHost = httplib.URL_IN
		msgHost = httplib.MSG_IN
	default:
		return errors.New("unsupported region; region must be one of (`CN`, `US`, `EU`, `IN`)")
	}

	connector.InitWithOptions(
		env.WithAccessID(d.serviceConfig.TuyaConnectorInfo.AccessId),
		env.WithAccessKey(d.serviceConfig.TuyaConnectorInfo.AccessKey),
		env.WithApiHost(apiHost),
		env.WithMsgHost(msgHost),
	)

	// we don't need sub message for this driver
	//go func() {
	//	message.Handler.InitMessageClient()
	//
	//	message.Handler.SubEventMessage(func(m *event.NameUpdateMessage) {
	//		d.Logger.Info("=========== name update： ==========")
	//		d.Logger.Infof("%v", *m)
	//	})
	//
	//	message.Handler.SubEventMessage(func(m *event.StatusReportMessage) {
	//		d.Logger.Info("=========== report data： ==========")
	//		for _, v := range m.Status {
	//			d.Logger.Infof(v.Code, v.Value)
	//		}
	//	})
	//}()

	return nil
}

func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	var (
		responses = make([]*sdkModel.CommandValue, len(reqs))
	)
	d.Logger.Debug(fmt.Sprintf("Driver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes))

	client, err := NewClientFromCache(deviceName, protocols)
	if err != nil {
		d.Logger.Error("Driver.HandleReadCommands: NewClient: %v", err)
		return nil, err
	}

	for i, req := range reqs {
		res, err := d.handleReadCommandRequest(client, req)
		if err != nil {
			d.Logger.Error(fmt.Sprintf("Driver.HandleReadCommands: Handle read commands failed: %v", err))
			return responses, err
		}
		responses[i] = res
	}

	return responses, nil
}

func (d *Driver) handleReadCommandRequest(client Client, req sdkModel.CommandRequest) (*sdkModel.CommandValue, error) {
	getParam, err := NewGetParam(req.Attributes)
	if err != nil {
		d.Logger.Error(fmt.Sprintf("Driver.handleReadCommands: NewGetParam failed: %s", err))
		return nil, err
	}
	reading, err := client.GetValue(getParam)
	if err != nil {
		d.Logger.Error(fmt.Sprintf("Driver.handleReadCommands: get value failed: %s", err))
		return nil, err
	}

	return NewCommandValue(req, reading)
}

func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	d.Logger.Debug(fmt.Sprintf("Driver.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[0].DeviceResourceName, params))
	var err error

	client, err := NewClientFromCache(deviceName, protocols)
	if err != nil {
		d.Logger.Error("Driver.HandleWriteCommands: NewClient: %v", err)
		return err
	}

	for i, req := range reqs {
		err := d.handleWriteCommandRequest(client, req, params[i])
		if err != nil {
			d.Logger.Info(fmt.Sprintf("Handle write commands failed: %v", err))
			return err
		}
	}

	return nil
}

func (d *Driver) handleWriteCommandRequest(client Client, req sdkModel.CommandRequest, param *sdkModel.CommandValue) error {
	reading, err := NewReading(req.Type, param)
	if err != nil {
		return err
	}

	setParam, err := NewSetParam(req.Attributes, reading)
	if err != nil {
		d.Logger.Error("Driver.handleWriteCommands: NewSetParam: %v", err)
		return err
	}
	err = client.SetValue(setParam)

	return err
}

func (d *Driver) Stop(force bool) error {
	d.cancel()
	d.clientMap.clients.For(func(key string, value interface{}) {
		client := value.(Client)
		client.Close()
	})
	return nil
}

func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.Logger.Warn("Driver's AddDevice function didn't implement")
	return nil
}

func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.Logger.Warn("Driver's UpdateDevice function didn't implement")
	return nil
}

func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.Logger.Warn("Driver's RemoveDevice function didn't implement")
	return nil
}
