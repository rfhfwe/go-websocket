package api

import (
	"encoding/json"
	"time"
)

// MsgInfo 接收的消息结构体
type MsgInfo struct {
	MessageType int        `json:"m"`
	TopicInfo   TopicInfo  `json:"t"`
	Context     MsgContext `json:"c"`
}

// TopicInfo 话题结构体
type TopicInfo struct {
	Topic string          `json:"topic"`
	Data  json.RawMessage `json:"data"`
	Type  string          `json:"type"`
}

// MsgContext 客户端推送的消息结构体
type MsgContext struct {
	ID        string            `json:"i"`
	StartTime time.Time         `json:"t"`
	ClientID  string            `json:"c"`
	UserID    string            `json:"u"`
	Source    MsgSource         `json:"s"`
	ExMeta    map[string]string `json:"e,omitempty"`
}

type PusherInfo interface {
	ClientID() string
	UserID() string
}

func NewMsgInfo(source MsgSource, pusher PusherInfo) *MsgInfo {
	ctx := MsgContext{
		Source:   source,
		ClientID: pusher.ClientID(),
		UserID:   pusher.UserID(),
	}
	m := &MsgInfo{
		TopicInfo: TopicInfo{Type: PublishKey},
		Context:   ctx,
	}
	return m
}
