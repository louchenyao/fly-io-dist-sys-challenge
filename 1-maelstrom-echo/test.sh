#! /bin/bash

# build
go build

# run test
maelstrom test -w echo --bin ./maelstrom-echo --node-count 1 --time-limit 10