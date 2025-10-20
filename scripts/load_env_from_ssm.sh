#!/usr/bin/env bash
# File: scripts/load_env_from_ssm.sh
# Purpose: Before `cdk synth/deploy`, load values from SSM and export env vars
# Usage: ENV=prod APP=user-votes REGION=us-east-2 ./scripts/load_env_from_ssm.sh
set -euo pipefail

: "${ENV:=prod}"
: "${APP:=user-votes}"
: "${REGION:=${AWS_REGION:-us-east-2}}"

PREFIX="/${APP}/${ENV}"

get_param() {
  local name="$1"; shift
  aws ssm get-parameter --name "$name" --with-decryption \
    --query Parameter.Value --output text --region "$REGION" 2>/dev/null || true
}

export COUNTERS_TABLE_NAME=$(get_param "$PREFIX/ddb/counters/name")
export ROMANCES_TABLE_NAME=$(get_param "$PREFIX/ddb/romances/name")
export DELETE_ROMANCES_TOPIC_ARN=$(get_param "$PREFIX/sns/delete-romances/arn")
export DELETE_ROMANCES_QUEUE_ARN=$(get_param "$PREFIX/sqs/delete-romances/arn")
export DELETE_ROMANCES_QUEUE_URL=$(get_param "$PREFIX/sqs/delete-romances/url")

# Friendly log
cat <<EOF
Loaded env from SSM (missing values left empty):
  COUNTERS_TABLE_NAME=$COUNTERS_TABLE_NAME
  ROMANCES_TABLE_NAME=$ROMANCES_TABLE_NAME
  DELETE_ROMANCES_TOPIC_ARN=$DELETE_ROMANCES_TOPIC_ARN
  DELETE_ROMANCES_QUEUE_ARN=$DELETE_ROMANCES_QUEUE_ARN
  DELETE_ROMANCES_QUEUE_URL=$DELETE_ROMANCES_QUEUE_URL
EOF