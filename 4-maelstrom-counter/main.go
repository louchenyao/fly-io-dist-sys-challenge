package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)

	type Value struct {
		Value   int
		Version int
	}

	add_delta := func(delta int) {
		ctx := context.Background()
		for {
			old := Value{0, 0}
			_ = kv.ReadInto(ctx, "counter", &old)
			new := Value{old.Value + delta, old.Version + 1}
			if err := kv.CompareAndSwap(ctx, "counter", old, new, true); err != nil {
				continue
			}
			_ = kv.Write(ctx, n.ID(), new) // update the local cache
			break
		}
	}

	read := func() Value {
		ctx := context.Background()
		global_v := Value{0, 0}
		local_v := Value{0, 0}
		_ = kv.ReadInto(ctx, "counter", &global_v)
		_ = kv.ReadInto(ctx, n.ID(), &local_v)
		// return the value with the highest version
		if local_v.Version > global_v.Version {
			return local_v
		}
		return global_v
	}

	send_all := func(v Value) {
		// broadcast the value to all nodes
		for _, node := range n.NodeIDs() {
			if node == n.ID() {
				continue
			}
			n.Send(node, map[string]any{
				"type":  "update",
				"value": v,
			})
		}
	}

	// broadcast the value to all nodes every second
	go func() {
		for {
			time.Sleep(1 * time.Second)
			send_all(read())
		}
	}()

	n.Handle("add", func(msg maelstrom.Message) error {
		message := struct {
			Delta int
		}{}
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		add_delta(message.Delta)

		return n.Reply(msg, map[string]any{
			"type": "add_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": read().Value,
		})
	})

	n.Handle("update", func(msg maelstrom.Message) error {
		message := struct {
			Value Value
		}{}
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		// update the local cache
		local := read()
		if message.Value.Version > local.Version {
			_ = kv.Write(context.Background(), n.ID(), message.Value)
		}

		return nil
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
