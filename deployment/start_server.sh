#!/usr/bin/env bash

set -ex

export PATH=$PATH:/usr/local/go/bin
export GO111MODULE=on
export GOPATH=/home/ec2-user/go
# export AWS_REGION=eu-central-1
# export INFURAID=$(aws secretsmanager get-secret-value --secret-id API_KEYS --query SecretString --region=us-east-1 --output text | jq -r .INFURAID)
# export BLOCKCHAIN_INFO_KEY=$(aws secretsmanager get-secret-value --secret-id API_KEYS --query SecretString --region=us-east-1 --output text | jq -r .BLOCKCHAIN_INFO_KEY)

service_name="herdius-core"
cd /home/ec2-user/go/src/github.com/herdius/herdius-core

supervisorctl stop "$service_name"

# Build server
make build-herserver

supervisorctl start "$service_name"
