#!/usr/bin/env bash
set -ex

LOGDIR=/var/log/herdius/herdius-core/log
export PATH=$PATH:/usr/local/go/bin
export GO111MODULE=on
export GOPATH=/home/ec2-user/go
cd /home/ec2-user/go/src/github.com/herdius/herdius-core
sudo mkdir -p $LOGDIR
sudo chmod 733 -R /var/log/herdius/
go get ./...
make start-supervisor ENV=staging > $LOGDIR/server.log 2> $LOGDIR/server.log < /dev/null &
echo "server started in background"