#! /bin/bash

go build

maelstrom test -w g-counter --bin ./maelstrom-counter --node-count 3 --rate 100 --time-limit 20 --nemesis partition