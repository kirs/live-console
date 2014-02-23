package main

import (
  "github.com/gorilla/websocket"
  "log"
  "fmt"
  "net/http"
)

type connection struct {
  ws *websocket.Conn
  // buffered channel of outbound messages
  send chan []byte
}

type sockethub struct {
  // registered connection
  connections map[*connection]bool
  // inbound messages from connections
  Broadcast chan []byte
  // register requests from connection
  register chan *connection
  // unregister request from connection
  unregister chan *connection
}

var H = sockethub{
  Broadcast:   make(chan []byte),
  register:    make(chan *connection),
  unregister:  make(chan *connection),
  connections: make(map[*connection]bool),
}

func (c *connection) writer() {
  for message := range c.send {
    err := c.ws.WriteMessage(1, message)
    if err != nil {
      log.Printf("Error in writer: ", err.Error())
      break
    }
  }
  c.ws.Close()
}

func PushHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  message := r.Form.Get("message")
  from := r.Form.Get("from")

  compiled := fmt.Sprintf("%s: %s", from, message)
  H.Broadcast <- []byte(compiled)
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
  ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
  if _, ok := err.(websocket.HandshakeError); ok {
    http.Error(w, "Not a websocket handshake", 400)
    return
  } else if err != nil {
    log.Printf("WsHandler error: ", err.Error())
    return
  }

  c := &connection{send: make(chan []byte, 256), ws: ws}
  H.register <- c
  //defer func() { H.unregister <- c }()
  c.writer()
}

func GithubHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("received:")
  fmt.Println(r.Body)
}

func (h *sockethub) Run() {
  for {
    select {
    case c := <-h.register:
      h.connections[c] = true
    case c := <-h.unregister:
      delete(h.connections, c)
      close(c.send)
    case m := <-h.Broadcast:
      for c := range h.connections {
        select {
        case c.send <- m:
          log.Printf("Broadcasting: %s", string(m))
        default:
          delete(h.connections, c)
          close(c.send)
          go c.ws.Close()
        }
      }
    }
  }
}


func main() {
  ws_host := "localhost:1235"

  go H.Run()
  http.HandleFunc("/ws", WsHandler)
  http.HandleFunc("/push", PushHandler)
  http.HandleFunc("/gh", GithubHandler)
  log.Println("Starting websocket server on: ", ws_host)
  http.ListenAndServe(ws_host, nil)
}
