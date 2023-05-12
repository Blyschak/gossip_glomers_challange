package main

import (
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	s := NewState()

	n.Handle("echo", func(msg maelstrom.Message) error { return s.HandleEcho(n, msg) })
	n.Handle("generate", func(msg maelstrom.Message) error { return s.HandleGenerate(n, msg) })
	n.Handle("broadcast", func(msg maelstrom.Message) error { return s.HandleBroadcast(n, msg) })
	n.Handle("read", func(msg maelstrom.Message) error { return s.HandleRead(n, msg) })
	n.Handle("topology", func(msg maelstrom.Message) error { return s.HandleTopology(n, msg) })

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
