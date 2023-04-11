package main

import (
	"context"
	"encoding/json"
	"log"

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
			break
		}
	}

	read := func() int {
		ctx := context.Background()
		v := Value{0, 0}
		_ = kv.ReadInto(ctx, "counter", &v)
		// it will return 0 if the key doesn't exist
		// it's ok since the system is eventually consistent
		return v.Value
	}

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
			"value": read(),
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
