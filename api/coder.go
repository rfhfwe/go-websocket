package api

import "github.com/vmihailenco/msgpack/v5"

type Coder interface {
	Decode([]byte, *MsgInfo) error
	Encode(msg *MsgInfo) []byte
}

type DefaultCoder struct{}

func (c *DefaultCoder) Decode(data []byte, msg *MsgInfo) error {
	return msgpack.Unmarshal(data, msg)
}

func (c *DefaultCoder) Encode(msg *MsgInfo) []byte {
	b, _ := msgpack.Marshal(msg)
	return b
}
