package driver

type GetResponse struct {
	Result []Command `json:"result"`
}

type SetResponse struct {
	Result bool `json:"result"`
}

type CommandReq struct {
	Commands []Command `json:"commands"`
}

type Command struct {
	Code  string `json:"code"`
	Value interface{}
}
