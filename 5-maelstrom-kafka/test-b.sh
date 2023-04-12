
#! /bin/bash

go build

maelstrom test -w kafka --bin ./maelstrom-kafka --node-count 2 --concurrency 2n --time-limit 20 --rate 1000