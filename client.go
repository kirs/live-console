package main

import "code.google.com/p/go.net/websocket"

import (
  "time"
  "github.com/wsxiaoys/terminal"
  "os/exec"
  // "github.com/wsxiaoys/terminal/color"
)

func sendToNc(message string) {
  cmd := exec.Command("terminal-notifier", "-message", message)
  cmd.Run()
}

func main() {
  ws, err := websocket.Dial("ws://localhost:1235/", "", "http://localhost:1235/")
  if err != nil {
    panic(err)
  }

  var resp = make([]byte, 4096)
  for {
    n, err := ws.Read(resp)
    if err != nil {
      panic(err)
    }

    var now, received string
    received = string(resp[0:n])

    now = time.Now().Format(time.RFC822)
    now += " "

    terminal.Stdout.Color("y").
        Print(now).Reset().Color("g").
        Print(received).Nl()

    sendToNc(received)
  }
}
