package main

import (
	"context"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c, _, err := websocket.Dial(ctx, "ws://localhost:2021/ws", nil)
	if err != nil {
		panic(err)
	}
	defer c.Close(websocket.StatusInternalError, "内部错误!")
	err = wsjson.Write(ctx, c, "Hello Websocket Server")
	if err != nil {
		panic(err)
	}
	var v interface{}
	err =
}