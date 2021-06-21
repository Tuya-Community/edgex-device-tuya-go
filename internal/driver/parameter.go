package driver

import (
	"fmt"
)

type BaseParam struct {
	Code string
}

type GetParam struct {
	BaseParam
}

type SetParam struct {
	BaseParam
	Value interface{}
}

func newBaseParam(attr map[string]interface{}) (base BaseParam, err error) {
	code, err := getStringFromAttr(Code, attr)
	if err != nil {
		return BaseParam{}, err
	}

	return BaseParam{
		Code: code,
	}, nil
}

func NewGetParam(attr map[string]interface{}) (*GetParam, error) {
	base, err := newBaseParam(attr)

	return &GetParam{
		BaseParam: base,
	}, err
}

func NewSetParam(attr map[string]interface{}, value interface{}) (*SetParam, error) {
	base, err := newBaseParam(attr)

	return &SetParam{
		BaseParam: base,
		Value:     value,
	}, err
}

func getStringFromAttr(fieldName string, attr map[string]interface{}) (string, error) {
	value, ok := attr[fieldName]
	if !ok {
		return "", fmt.Errorf("can't fetch `%s`", fieldName)
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("`%s` must be string", fieldName)
	}

	return valueStr, nil
}
