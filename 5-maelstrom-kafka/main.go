package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func send(kv *maelstrom.KV, key string, msg any) (int, error) {
	ctx := context.Background()

	// fetch and increment the offset
	offset := -1
	for {
		if err := kv.ReadInto(ctx, key+"/offset", &offset); err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
			return -1, err
		}
		if err := kv.CompareAndSwap(ctx, key+"/offset", offset, offset+1, true); err == nil {
			break
		}
	}
	offset += 1

	// if the system crashes here, we'll have a gap in the sequence of offsets

	// write the message
	if err := kv.Write(ctx, key+"/"+fmt.Sprintf("%d", offset), msg); err != nil {
		return -1, err
	}

	return offset, nil
}

// poll returns a slice of tuples of the form [offset, message]
func poll(kv *maelstrom.KV, key string, start int) ([][]any, error) {
	ctx := context.Background()
	latest_offset := -1
	if err := kv.ReadInto(ctx, key+"/offset", &latest_offset); err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
		return nil, err
	}
	limit := 100 // arbitrary
	msgs := [][]any{}
	for i := start; i <= start+limit && i <= latest_offset; i++ {
		msg, err := kv.Read(ctx, key+"/"+fmt.Sprintf("%d", i))
		if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
			continue
		} else if err != nil {
			return nil, err
		}
		msgs = append(msgs, []any{any(i), msg})
	}
	return msgs, nil
}

func commit_offsets(kv *maelstrom.KV, key string, offset int) error {
	ctx := context.Background()
	return kv.Write(ctx, key+"/committed_offset", offset)
}

// list_committed_offsets returns the last committed offset for a key, or -1 if none exists
func list_committed_offsets(kv *maelstrom.KV, key string) (int, error) {
	ctx := context.Background()
	offset := -1
	if err := kv.ReadInto(ctx, key+"/committed_offset", &offset); err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
		return -1, err
	}
	return offset, nil
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewLinKV(n)

	n.Handle("send", func(msg maelstrom.Message) error {
		message := struct {
			Key string
			Msg any
		}{}
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		offset, err := send(kv, message.Key, message.Msg)
		if err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{
			"type":   "send_ok",
			"offset": offset,
		})
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		message := struct {
			Offsets map[string]int
		}{}
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		msgs := map[string][][]any{}
		for key, offset := range message.Offsets {
			m, err := poll(kv, key, offset)
			if err != nil {
				return err
			}
			msgs[key] = m
		}

		return n.Reply(msg, map[string]any{
			"type": "poll_ok",
			"msgs": msgs,
		})
	})

	n.Handle("commit_offsets", func(msg maelstrom.Message) error {
		message := struct {
			Offsets map[string]int
		}{}
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		for key, offset := range message.Offsets {
			if err := commit_offsets(kv, key, offset); err != nil {
				return err
			}
		}

		return n.Reply(msg, map[string]any{
			"type": "commit_offsets_ok",
		})
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		message := struct {
			Keys []string
		}{}
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		offsets := map[string]int{}
		for _, key := range message.Keys {
			offset, err := list_committed_offsets(kv, key)
			if err != nil {
				return err
			}
			// -1 means no committed offset
			if offset != -1 {
				offsets[key] = offset
			}
		}

		return n.Reply(msg, map[string]any{
			"type":    "list_committed_offsets_ok",
			"offsets": offsets,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
