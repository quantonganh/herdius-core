#!/usr/bin/env bash
set -ex

export PATH=$PATH:/usr/local/go/bin
export GOPATH=/home/ec2-user/go
whoami
ls -lah
pwd
cd /home/ec2-user/go/src/github.com/herdius/herdius-core
go get ./...
make start-server ENV=staging > /dev/null 2> /dev/null < /dev/null &
echo "server started in background"
