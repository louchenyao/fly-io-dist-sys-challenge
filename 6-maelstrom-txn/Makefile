build:
	go build

test-a: build
	maelstrom test -w txn-rw-register --bin ./maelstrom-txn --node-count 1 --time-limit 20 --rate 1000 --concurrency 2n --consistency-models read-uncommitted --availability total

test-b-1: build
	maelstrom test -w txn-rw-register --bin ./maelstrom-txn --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-uncommitted

test-b-2: build
	maelstrom test -w txn-rw-register --bin ./maelstrom-txn --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-uncommitted --availability total --nemesis partition

test-c: build
	maelstrom test -w txn-rw-register --bin ./maelstrom-txn --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-committed --availability total --nemesis partition

.PHONY: build