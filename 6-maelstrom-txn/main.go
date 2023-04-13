package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func read(kv *maelstrom.KV, key string) (any, error) {
	ctx := context.Background()
	if val, err := kv.Read(ctx, key); err == nil {
		return val, nil
	} else if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
		return nil, nil
	} else {
		return nil, err
	}
}

func write(kv *maelstrom.KV, key string, val any) error {
	ctx := context.Background()
	return kv.Write(ctx, key, val)
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewLinKV(n)

	n.Handle("txn", func(msg maelstrom.Message) error {
		message := struct {
			Txn [][]any
		}{}
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		write_set := map[string]any{}
		var err error
		for i, op := range message.Txn {
			key := fmt.Sprintf("%v", op[1])
			switch op[0] {
			case "r":
				if val, ok := write_set[key]; ok {
					message.Txn[i][2] = val
				} else {
					message.Txn[i][2], err = read(kv, key)
					if err != nil {
						return err
					}
				}
			case "w":
				write_set[key] = op[2]
			}
		}
		// apply writes
		for key, val := range write_set {
			err = write(kv, key, val)
			if err != nil {
				return err
			}
		}

		return n.Reply(msg, map[string]any{
			"type": "txn_ok",
			"txn":  message.Txn,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
