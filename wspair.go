package wspipe

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrConflicted = errors.New("conflicted")

type wspair struct {
	Left       *websocket.Conn
	Right      *websocket.Conn
	sync.Mutex // 保护两端连接

	ready     chan struct{}
	sync.Once // 保护 ready 信号
}

func (self *wspair) Close() { self.broadcast() }

func (self *wspair) Join(conn *websocket.Conn) (bool, error) {
	var right bool
	var err error
	self.Lock()
	if self.Left == nil {
		self.Left = conn
		right = false
	} else if self.Right == nil {
		self.Right = conn
		right = true
	} else {
		err = ErrConflicted
	}
	self.Unlock()
	if right && err == nil {
		self.broadcast()
	}
	return right, err
}

func (self *wspair) Ready() <-chan struct{} { return self.ready }
func (self *wspair) broadcast()             { self.Do(func() { close(self.ready) }) }
