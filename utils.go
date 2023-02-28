package wspipe

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

func wscopy(ctx context.Context, left, right *websocket.Conn) {
	running := true
	done := ctx.Done()
	for running {
		select {
		case <-done:
			running = false
		default:
			if tpe, msg, err := right.ReadMessage(); err != nil {
				left.SetReadDeadline(time.Now())
				running = false
			} else if err = left.WriteMessage(tpe, msg); err != nil {
				right.SetWriteDeadline(time.Now())
				running = false
			}
		}
	}
}
