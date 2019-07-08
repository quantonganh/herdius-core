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
  chmod 733 -R "$LOGDIR"
fi


if [[ ! -d "$RUNDIR" ]]; then
  mkdir -p "$RUNDIR"
  chmod 733 -R "$RUNDIR"
fi

# Build server
make build-herserver

# Start supervisor or validator base on $type
if [ "$type" = "supervisor" ]; then
  ./herserver -supervisor=true -groupsize=3 -port=0 -waitTime=15 -env=staging >"$logfile" 2>&1 </dev/null &
else
  ./herserver -peers="tcp://127.0.0.1:3000" -groupsize=3 -port=3001 -waitTime=15 -env=staging >"$logfile" 2>&1 </dev/null &
fi

# Save the pid to kill later
printf '%s\n' "$!" >"$pidfile"

echo "${type} started in background"
