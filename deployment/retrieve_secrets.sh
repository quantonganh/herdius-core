#!/usr/bin/env bash
set -ex

export INFURAID=$(aws secretsmanager get-secret-value --secret-id INFURAID --query SecretString --region=us-east-1 --output text | jq -r .INFURAID)
echo "INFURAID VALUE IS:"
echo "PWD: $PWD"
echo "$INFURAID" >> /home/ec2-user/go/src/github.com/herdius/herdius-core/env_vars.log