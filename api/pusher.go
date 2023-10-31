package api

import "sync"

type Pusher interface {
	Publish(msg *MsgInfo) error
	Receive() chan *MsgInfo
}
type SinglePusher struct {
	msg     chan []byte
	once    sync.Once
	coder   Coder
	msgChan chan *MsgInfo
}

func (s *SinglePusher) Pusher(msg *MsgInfo) error {
	s.msg <- s.coder.Encode(msg)
	return nil
}

func (s *SinglePusher) Receive() chan *MsgInfo {
	s.once.Do(func() {
		for {
			select {
			case m := <-s.msg:
				msg := new(MsgInfo)
				err := s.coder.Decode(m, msg)
				if err != nil {
					continue
				}
				s.msgChan <- msg
			}
		}
	})
	return s.msgChan
}

func DefaultPusher(coder Coder) *SinglePusher {
	return &SinglePusher{
		msg:     make(chan []byte, 100),
		coder:   coder,
		msgChan: make(chan *MsgInfo),
	}
}
