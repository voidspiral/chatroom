package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
)
type webhandler1 struct {

}
func (webhandler1) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	writer.Write([]byte("web1"))
}
type webhandler2 struct {

}

func (webhandler2) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	writer.Write([]byte("web2"))
}
func main() {

	c := make(chan os.Signal)
	go (func(){
		http.ListenAndServe("9091", webhandler1{})
	})()
	go (func() {
		http.ListenAndServe("9092", webhandler2{})
	})()
	signal.Notify(c, os.Interrupt)
	s := <- c
	log.Println(s)
}
