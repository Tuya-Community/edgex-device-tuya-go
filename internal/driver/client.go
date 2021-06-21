package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/tuya/tuya-connector-go/connector"
	"net/http"
	"time"
)

type Client interface {
	Ping() bool
	Close()
	GetValue(param *GetParam) (interface{}, error)
	SetValue(param *SetParam) error
}

type tuyaClient struct {
	getUriApi string
	setUriApi string
}

func NewClient(info *ConnectionInfo) (Client, error) {
	return &tuyaClient{
		getUriApi: fmt.Sprintf(GetUriFormat, info.DeviceId),
		setUriApi: fmt.Sprintf(SetUriFormat, info.DeviceId),
	}, nil
}

func (c *tuyaClient) Ping() bool {

	return true
}

func (c *tuyaClient) Close() {

	return
}

func (c *tuyaClient) GetValue(param *GetParam) (interface{}, error) {
	resp := &GetResponse{}
	err := connector.MakeRequest(
		context.Background(),
		connector.WithMethod(http.MethodGet),
		connector.WithAPIUri(c.getUriApi),
		connector.WithResp(resp),
	)

	for _, each := range resp.Result {
		if each.Code == param.Code {
			return each.Value, nil
		}
	}

	return nil, err
}

func (c *tuyaClient) SetValue(param *SetParam) error {
	cmdReq := CommandReq{Commands: []Command{
		{
			Code:  param.Code,
			Value: param.Value,
		},
	}}
	bts, err := json.Marshal(cmdReq)
	resp := &SetResponse{}
	err = connector.MakeRequest(
		context.Background(),
		connector.WithMethod(http.MethodPost),
		connector.WithAPIUri(c.setUriApi),
		connector.WithPayload(bts),
		connector.WithResp(resp),
	)

	if !resp.Result {
		return errors.New("send set command error")
	}

	return err
}

func NewClientFromCache(deviceId string, protocols map[string]models.ProtocolProperties) (Client, error) {
	client, ok := driver.clientMap.Load(deviceId)
	if !ok || client == nil || !client.Ping() {
		info, err := CreateConnectionInfo(protocols)
		if err != nil {
			return nil, err
		}

		client, err = NewClient(info)
		if err != nil {
			return nil, err
		}
		driver.clientMap.Store(deviceId, client)
	} else {
		driver.clientMap.Grow(deviceId, 3*time.Minute)
	}

	return client, nil
}
