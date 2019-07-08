#!/usr/bin/env bash

set -ex

export LOGDIR=/var/log/herdius/herdius-core/log
export RUNDIR=/var/run/herdius
export PATH=$PATH:/usr/local/go/bin
export GO111MODULE=on
export GOPATH=/home/ec2-user/go
export AWS_REGION=eu-central-1
export INFURAID=$(aws secretsmanager get-secret-value --secret-id API_KEYS --query SecretString --region=us-east-1 --output text | jq -r .INFURAID)
export BLOCKCHAIN_INFO_KEY=$(aws secretsmanager get-secret-value --secret-id API_KEYS --query SecretString --region=us-east-1 --output text | jq -r .BLOCKCHAIN_INFO_KEY)

usage() {
  echo "Usage: ${0##*/} [supervisor|validator]"
  exit 1
}

type=${1:-"supervisor"}
pidfile=""
logfile=""

case "$type" in
  (supervisor)
               pidfile="${RUNDIR}/supervisor.pid"
               logfile="${LOGDIR}/supervisor.log"
               ;;
   (validator)
               pidfile="${RUNDIR}/validator.pid"
               logfile="${LOGDIR}/validator.log"
               PORT=3001
               PEERS="tcp://127.0.0.1:3000"
               HOST="127.0.0.1"
               ;;
           (*)
               usage
               ;;
esac

# Kill old process if existed
if [[ -f "$pidfile" ]]; then
  kill "$(cat "$pidfile")" || :
fi

# Preparation
cd /home/ec2-user/go/src/github.com/herdius/herdius-core

if [[ ! -d "$LOGDIR" ]]; then
  mkdir -p "$LOGDIR"
fi

chmod 733 -R /var/log/herdius/

if [[ ! -d "$RUNDIR" ]]; then
  mkdir -p "$RUNDIR"
fi

# Start supervisor or validator base on $type
make start-"$type" ENV=staging PORT="$PORT" PEERS="$PEERS" HOST="$HOST" >"$logfile" 2>&1 </dev/null &

# Save the pid to kill later
printf '%s\n' "$!" >"$pidfile"

echo "${type} started in background"
