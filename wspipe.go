package wspipe

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WSPipe struct {
	upgrader websocket.Upgrader
	pairs    sync.Map
	version  string
	token    string
}

func (self *WSPipe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	} else if self.token != r.URL.Query().Get("token") {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	} else if conn, err := self.upgrader.Upgrade(w, r, nil); err == nil {
		defer conn.Close()
		path := r.URL.Path
		pair := self.getPair(path)
		if right, err := pair.Join(conn); err == nil {
			defer self.deletePair(path)
			defer pair.Close()
			if right {
				wscopy(r.Context(), pair.Left, conn)
			} else {
				ctx := r.Context()
				select {
				case <-ctx.Done():
				case <-pair.Ready():
					wscopy(ctx, pair.Right, conn)
				}
			}
		}
	}
}

func (self *WSPipe) getPair(path string) *wspair {
	pair := &wspair{ready: make(chan struct{})}
	if actual, loaded := self.pairs.LoadOrStore(path, pair); loaded {
		close(pair.ready)
		return actual.(*wspair)
	} else {
		return actual.(*wspair)
	}
}

func (self *WSPipe) deletePair(path string) { self.pairs.Delete(path) }

func Build(ver, token string) (http.Handler, error) {
	return &WSPipe{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(*http.Request) bool { return true },
		},
		version: ver,
		token:   token,
	}, nil
}
