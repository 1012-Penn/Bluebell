#!/bin/bash
set -e

go build -o ./bin/bluebell
./bin/bluebell conf/config.yaml &
echo $! > ./bin/bluebell.pid 