// This tool reads newline-delimited data from stdin and pumps it out to a
// WebSocket server.
package main

import (
	"bufio"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	url := "ws://localhost:8888/input"
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal(err)
	}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		w, err := c.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(s.Bytes())
		w.Close()
	}
}
