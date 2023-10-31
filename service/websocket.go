package post

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go-websocket/api"
	"go-websocket/pkg/errorinfo"
	"go-websocket/utils"
	"strings"
	"time"
)

// Post 客户端连接的结构体
type Post struct {
	topic     map[string]bool // 订阅topic列表
	clientID  string          // 客户端ID
	userId    string          // 用户Id
	isClose   bool            // 判断是否被关闭
	closeChan chan struct{}   // websocket关闭的信号

	readIn  chan *api.MsgInfo
	sendOut chan *api.MsgInfo

	ws *websocket.Conn // 底层的ws连接

	onConnectHandler       func() bool
	onOfflineHandler       func()
	receivedHandler        func() bool
	readHandler            func() bool
	readTimeoutHandler     func()
	subscribeHandler       func() bool
	unsubscribeHandler     func() bool
	beforeSubscribeHandler func() bool
	onSystemRemove         func(topic string)
}

func (p *Post) ClientID() string {
	return p.clientID
}

func (p *Post) UserID() string {
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

// OnClose 返回Post的状态
func (p *Post) OnClose() chan struct{} {
	return p.closeChan
}

// Run 启动websocket
func (p *Post) Run() {
	// 读取websocket信息
	go p.readLoop()
	// 处理读取事件
	go p.readDispose()

	if p.onConnectHandler != nil {
		ok := p.onConnectHandler()
		if !ok {
			p.Close()
		}
	}
	// 向websocket发送消息
	p.sendLoop()
}

// Subscribe 订阅
func (p *Post) Subscribe(ctx api.MsgContext, topics []string) error {
	if p.beforeSubscribeHandler != nil {
		ok := p.beforeSubscribeHandler() // 触发订阅前事件
		if !ok {
			return nil
		}
	}
	// topic 绑定
	_, err := p.bindTopic(topics)
	if err != nil {
		return err
	}
	if p.subscribeHandler != nil {
		p.subscribeHandler() // 触发订阅时事件
	}
	return nil
}

// UnSubscribe 取消订阅
func (p *Post) UnSubscribe(ctx api.MsgContext, topics []string) error {
	// 解绑
	_, err := p.unbindTopic(topics)
	if err != nil {
		return err
	}
	if p.unsubscribeHandler != nil {
		p.unsubscribeHandler()
	}
	return nil
}

// Publish 发布消息
func (p *Post) Publish(msg *api.MsgInfo) error {
	// 推送消息出去
	err := pm.Publish(msg)
	if err != nil {
		return err
	}
	return nil
}

// Close 关闭客户端连接
func (p *Post) Close() {
	if !p.isClose {
		p.isClose = true
		if p.topic != nil {
			// 在manager下解绑我订阅的topic
			var topics []string
			for k := range p.topic {
				topics = append(topics, k)
			}
			_, err := p.unbindTopic(topics)
			if err != nil {
				// todo: 打log
			}

			if p.unsubscribeHandler != nil { // 执行回调方法
				p.unsubscribeHandler()
			}
		}
		_ = p.ws.Close()
		if p.onOfflineHandler != nil {
			p.onOfflineHandler()
		}
		close(p.closeChan)
	}
}

// TopicList 返回用户订阅的所有topic
func (p *Post) TopicList() []string {
	var topics []string = make([]string, 0, len(p.topic))
	for k := range p.topic {
		topics = append(topics, k)
	}
	return topics
}

// bindTopic topic绑定
func (p *Post) bindTopic(topic []string) ([]string, error) {
	// 放到manager中对应的map中
	var (
		addTopic []string
	)
	for _, v := range topic {
		if _, ok := p.topic[v]; !ok {
			addTopic = append(addTopic, v)
			p.topic[v] = true
			pm.AddSubscribe(v, p)
		}
	}
	return addTopic, nil
}

// unbindTopic 解绑topic
func (p *Post) unbindTopic(topic []string) ([]string, error) {
	// 从manager对应的map中移除
	var (
		delTopic []string
	)
	for _, v := range topic {
		if _, ok := p.topic[v]; !ok {
			delTopic = append(delTopic, v)
			p.topic[v] = true
			pm.DelSubscribe(v, p)
		}
	}
	return delTopic, nil
}

// SetOnConnectHandler 用户连接时触发
func (p *Post) SetOnConnectHandler(fn func() bool) {
	p.onConnectHandler = fn
}

// SetOnOfflineHandler 用户关闭时触发
func (p *Post) SetOnOfflineHandler(fn func()) {
	p.onOfflineHandler = fn
}

// SetReceivedHandler 向用户推送消息时触发
func (p *Post) SetReceivedHandler(fn func() bool) {
	p.receivedHandler = fn
}

// SetReadHandler 接收到用户publish的消息时触发
func (p *Post) SetReadHandler(fn func() bool) {
	p.readHandler = fn
}

// SetSubscribeHandler 用户订阅topic后触发
func (p *Post) SetSubscribeHandler(fn func() bool) {
	p.subscribeHandler = fn
}

// SetBeforeSubscribeHandler 用户订阅topic前触发
func (p *Post) SetBeforeSubscribeHandler(fn func() bool) {
	p.beforeSubscribeHandler = fn
}

// SetReadTimeoutHandler 生产 > 消费时触发
func (p *Post) SetReadTimeoutHandler(fn func()) {
	p.readTimeoutHandler = fn
}

// SetSystemRemove 移除系统中某个用户的topic订阅
func (p *Post) SetSystemRemove(fn func(topic string)) {
	p.onSystemRemove = fn
}

// readLoop 读取ws连接中的信息
func (p *Post) readLoop() {

	for {
		messageType, data, err := p.ws.ReadMessage()
		if err != nil {
			goto collapse
		}
		msg := api.NewMsgInfo(api.SourceClient, p)
		msg.MessageType = messageType
		if err = json.Unmarshal(data, &msg.TopicInfo); err != nil {
			continue
		}
		// 设置超时时间
		timeout := time.After(time.Duration(3) * time.Second)
		select {
		case p.readIn <- msg: // 将消息传到读取队列
		case <-timeout:
			if p.readTimeoutHandler != nil {
				p.readTimeoutHandler()
			}
		case <-p.closeChan:
			return
		}
	}

collapse:
	p.Close()
}

// readDispose 处理读取事件，判断用户是要进行通信还是topic订阅
func (p *Post) readDispose() {
	for {
		msg, err := p.read()
		if err != nil {
			p.Close()
			return
		}
		if p.isClose {
			return
		}
		if utils.IsBlank(msg.TopicInfo.Topic) {
			continue
		}
		switch msg.TopicInfo.Type {
		case api.Subscribe: // 用户订阅topic
			addTopics := strings.Split(msg.TopicInfo.Topic, ",")
			err = p.Subscribe(msg.Context, addTopics)
			if err != nil {
				// todo 打一个log
			}
		case api.UnSubscribe: // 用户取消订阅topic
			delTopics := strings.Split(msg.TopicInfo.Topic, ",")
			err = p.UnSubscribe(msg.Context, delTopics)
			if err != nil {
				// todo 打一个log
			}
		default: // 默认为用户发送消息
			if p.readHandler != nil {
				ok := p.readHandler()
				if !ok {
					p.Close()
					return
				}
			}
		}
	}
}

// sendLoop 向websocket发送信息
func (p *Post) sendLoop() {
	for {
		select {
		case msg := <-p.sendOut:
			if msg.MessageType == 0 {
				msg.MessageType = 1 // 文本格式
			}
			if err := p.ws.WriteMessage(msg.MessageType, []byte(msg.TopicInfo.Data)); err != nil {
				goto collapse
			}
		case <-p.closeChan:
			return
		}
	}
collapse:
	p.Close()
}

func (p *Post) read() (*api.MsgInfo, error) {
	if p.isClose {
		return nil, errorinfo.ErrorClose
	}
	return <-p.readIn, nil
}
