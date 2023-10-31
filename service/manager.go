package post

import "go-websocket/api"

var (
	pm *PostManager
)

type Manager interface {
	api.Pusher
	GetTopics() (map[string]uint64, error)
}

// PostManager ws管理端
type PostManager struct {
	posts map[string]map[string]*Post // topic -> clientID -> ws.conn

	isClose   bool
	closeChan chan struct{}

	coder api.Coder
	api.Pusher
}

func BuildFoundation() (Manager, error) {
	pm := &PostManager{}
	// todo
	return pm, nil
}

func (pm *PostManager) GetTopics() (map[string]uint64, error) {
	return nil, nil
}

func (pm *PostManager) AddSubscribe(topic string, post *Post) {
	// todo
}

func (pm *PostManager) DelSubscribe(topic string, post *Post) {
	// todo
}

func (pm *PostManager) Publish(msg *api.MsgInfo) error {
	// todo
	return nil
}
