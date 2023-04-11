package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	messages := make(map[int64]bool)
	lock := sync.RWMutex{}

	broadcast := func(value int64) {
		for _, dest := range n.NodeIDs() {
			n.RPC(dest, map[string]any{"type": "broadcast", "message": value}, func(msg maelstrom.Message) error {
				// ignore the response
				return nil
			})
		}
	}

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		message := struct {
			Message int64
		}{}
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}
		value := message.Message

		lock.RLock()
		_, ok := messages[value]
		lock.RUnlock()

		if ok {
			return nil
		} else {
			lock.Lock()
			messages[value] = true
			lock.Unlock()
			broadcast(value)
		}

		return n.Reply(msg, map[string]string{
			"type": "broadcast_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		// flatten the map into a slice of keys
		lock.RLock()
		keys := make([]int64, 0, len(messages))
		for k := range messages {
			keys = append(keys, k)
		}
		lock.RUnlock()

		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": keys,
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		// since we broadcast to all nodes, we don't care about the topology
		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
