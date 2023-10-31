package post

import (
	"github.com/gorilla/websocket"
)

// Post 客户端连接的结构体
type Post struct {
	ws        *websocket.Conn // 底层的ws连接
	topic     map[string]bool // 订阅topic列表
	clientID  string          // 客户端ID
	userId    string          // 用户Id
	isClose   bool            // 判断是否被关闭
	closeChan chan struct{}   // websocket关闭的信号
}

func (p *Post) GetClient() string {
	return p.clientID
}

func (p *Post) getUserID() string {
	return p.userId
}

func (p *Post) setUserID(id string) {
	p.userId = id
}

func NewPost(ws *websocket.Conn, clientId string) *Post {
	if ws == nil {
		return nil
	}
	post := &Post{
		clientID:  clientId,
		topic:     make(map[string]bool),
		ws:        ws,
		closeChan: make(chan struct{}),
	}
	return post
}

func (p *Post) OnClose() chan struct{} {
	return p.closeChan
}

// Run 启动websocket
func (p *Post) Run() {
	// TODO 读取websocket信息
	// TODO 处理读取事件
	// TODO 向websocket发送消息
}

// Subscribe 订阅
func (p *Post) Subscribe() error {
	// TODO
}

// UnSubscribe 取消订阅
func (p *Post) UnSubscribe() error {
	// TODO
}

// Publish 发布消息
func (p *Post) Publish() error {
	// todo
}

// Close 关闭客户端连接
func (p *Post) Close() {
	if !p.isClose {
		p.isClose = true
		if p.topic != nil {
			// TODO
		}
		_ = p.ws.Close()

		close(p.closeChan)
	}
}
