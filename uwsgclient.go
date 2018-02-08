package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type uwsclient struct {
	interrupt chan os.Signal
	done      chan struct{}
	url       url.URL
	conn      *websocket.Conn
}

var addr = "localhost:15051"
var instance *uwsclient
var once sync.Once

// GetInstanceUWS Singleton UWSClient
func GetInstanceUWS() *uwsclient {
	fmt.Println("=================== uwsclient.GetInstanceUWS ===================")
	once.Do(func() {
		var uws uwsclient
		var err error
		uws, err = InitUWS()
		if err != nil {
			fmt.Println("[xxxxxx] ERROR:", err)
			//log.Fatal("dial:", err)
		}
		instance = &uws
	})
	return instance
}

// InitUWS UWSClient
func InitUWS() (uwsclient, error) {
	var uws uwsclient
	var err error
	uws.interrupt = make(chan os.Signal, 1)
	signal.Notify(uws.interrupt, os.Interrupt)
	uws.done = make(chan struct{})

	uws.url = url.URL{Scheme: "ws", Host: addr, Path: "/"}
	log.Printf("connecting to %s", uws.url.String())

	uws.conn, _, err = websocket.DefaultDialer.Dial(uws.url.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	return uws, err
}

func (uws uwsclient) Close() {
	if uws.conn != nil {
		uws.conn.Close()
	}
}

func main() {
	var uws *uwsclient
	// var err error
	uws = GetInstanceUWS()
	defer uws.Close()

	// Thread receive message.
	go func() {
		defer uws.Close()
		defer close(uws.done)
		for {
			_, message, err := uws.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	// Thread send message.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			err := uws.conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-uws.interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := uws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-uws.done:
			case <-time.After(time.Second):
			}
			uws.Close()
			return
		}
	}
}

// func main() {
// 	var uws uwsclient
// 	var err error
// 	uws, err = InitUWS()
// 	if err != nil {
// 		log.Fatal("dial:", err)
// 	}
// 	defer uws.Close()
// 	// Thread receive message.
// 	go func() {
// 		defer uws.Close()
// 		defer close(uws.done)
// 		for {
// 			_, message, err := uws.conn.ReadMessage()
// 			if err != nil {
// 				log.Println("read:", err)
// 				return
// 			}
// 			log.Printf("recv: %s", message)
// 		}
// 	}()
// 	// Thread send message.
// 	ticker := time.NewTicker(time.Second)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case t := <-ticker.C:
// 			err := uws.conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
// 			if err != nil {
// 				log.Println("write:", err)
// 				return
// 			}
// 		case <-uws.interrupt:
// 			log.Println("interrupt")
// 			// To cleanly close a connection, a client should send a close
// 			// frame and wait for the server to close the connection.
// 			err := uws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
// 			if err != nil {
// 				log.Println("write close:", err)
// 				return
// 			}
// 			select {
// 			case <-uws.done:
// 			case <-time.After(time.Second):
// 			}
// 			uws.Close()
// 			return
// 		}
// 	}
// }

// func main() {
// 	interrupt := make(chan os.Signal, 1)
// 	signal.Notify(interrupt, os.Interrupt)
// 	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
// 	log.Printf("connecting to %s", u.String())
// 	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
// 	if err != nil {
// 		log.Fatal("dial:", err)
// 	}
// 	defer c.Close()
// 	done := make(chan struct{})
// 	// Thread receive message.
// 	go func() {
// 		defer c.Close()
// 		defer close(done)
// 		for {
// 			_, message, err := c.ReadMessage()
// 			if err != nil {
// 				log.Println("read:", err)
// 				return
// 			}
// 			log.Printf("recv: %s", message)
// 		}
// 	}()
// 	// Thread send message.
// 	ticker := time.NewTicker(time.Second)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case t := <-ticker.C:
// 			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
// 			if err != nil {
// 				log.Println("write:", err)
// 				return
// 			}
// 		case <-interrupt:
// 			log.Println("interrupt")
// 			// To cleanly close a connection, a client should send a close
// 			// frame and wait for the server to close the connection.
// 			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
// 			if err != nil {
// 				log.Println("write close:", err)
// 				return
// 			}
// 			select {
// 			case <-done:
// 			case <-time.After(time.Second):
// 			}
// 			c.Close()
// 			return
// 		}
// 	}
// }
