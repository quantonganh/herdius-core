#!/usr/bin/env bash
set -ex

export LOGDIR=/var/log/herdius/herdius-core/log
export PATH=$PATH:/usr/local/go/bin
export GO111MODULE=on
export GOPATH=/home/ec2-user/go
export AWS_REGION=eu-central-1
cd /home/ec2-user/go/src/github.com/herdius/herdius-core
sudo mkdir -p $LOGDIR
sudo chmod 733 -R /var/log/herdius/
go get ./...

export INFURAID=$(aws secretsmanager get-secret-value --secret-id INFURAID --query SecretString --region=us-east-1 --output text | jq -r .INFURAID)
echo "INFURAID VALUE IS:"
echo "PWD: $PWD"
echo "$INFURAID" >> /home/ec2-user/go/src/github.com/herdius/herdius-core/env_vars.log

make start-supervisor ENV=staging > $LOGDIR/server.log 2> $LOGDIR/server.log < /dev/null &
echo "server started in background"