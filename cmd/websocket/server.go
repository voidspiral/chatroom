package main

import (
	"context"
	"fmt"
	"log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"

	"net/http"
)

func main() {
	//a, b := net.Dial("tcp", "localhost")
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println(w, "HTTP, Hello")
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request){
		conn, err := websocket.Accept(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close(websocket.StatusInternalError, "内部出错了")
		ctx, cancel := context.WithTimeout(req.Context(), time.Second * 10)
		defer cancel()

		var v interface{}
		err = wsjson.Read(ctx, conn, &v)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("接收到客户端:%v\n", v)
		err = wsjson.Write(ctx, conn, "Hello Websocket Client")
		if err != nil {
			log.Println(err)
			return
		}
		conn.Close(websocket.StatusNormalClosure, "")
	})
	log.Fatal(http.ListenAndServe(":2021", nil))
}