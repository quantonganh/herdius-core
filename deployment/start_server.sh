#!/usr/bin/env bash
set -ex

export PATH=$PATH:/usr/local/go/bin
export GO111MODULE=on
export GOPATH=/home/ec2-user/go
cd /home/ec2-user/go/src/github.com/herdius/herdius-core
mkdir -p /var/log/herdius/herdius-core/log/
sudo chmod 733 -R /var/log/herdius/
go get ./...
make start-supervisor ENV=staging > /var/log/herdius/herdius-core/log/server.log 2> /var/log/herdius/herdius-core/log/server.log < /dev/null &
echo "server started in background"
