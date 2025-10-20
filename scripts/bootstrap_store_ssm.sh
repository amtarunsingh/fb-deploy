#!/usr/bin/env bash
# File: scripts/bootstrap_store_ssm.sh
# Purpose: After first `cdk deploy`, capture stack outputs and persist them to SSM Parameter Store
# Usage: ENV=prod APP=user-votes REGION=us-east-2 ACCOUNT=123456789012 ./scripts/bootstrap_store_ssm.sh
set -euo pipefail

: "${ENV:=prod}"
: "${APP:=user-votes}"
: "${REGION:=${AWS_REGION:-us-east-2}}"
: "${ACCOUNT:=${AWS_ACCOUNT_ID:-${CDK_DEFAULT_ACCOUNT:-}}}"

STACK_DATA=DataStack
STACK_SERVICE=ServiceStack

# Helper: put a String parameter (idempotent with --overwrite)
put_param() {
  local name="$1"; shift
  local value="$1"; shift
  aws ssm put-parameter \
    --name "$name" \
    --value "$value" \
    --type String \
    --overwrite \
    --region "$REGION" 1>/dev/null
  echo "Saved $name => $value"
}

# Pull CloudFormation outputs by key
cf_output() {
  local stack="$1"; shift
  local key="$1"; shift
  aws cloudformation describe-stacks \
    --stack-name "$stack" \
    --query "Stacks[0].Outputs[?OutputKey==\``$key`\`].OutputValue" \
    --output text \
    --region "$REGION"
}

# 1) Read outputs from DataStack
COUNTERS_NAME=$(cf_output "$STACK_DATA" CountersTableName || true)
ROMANCES_NAME=$(cf_output "$STACK_DATA" RomancesTableName || true)
TOPIC_ARN=$(cf_output "$STACK_DATA" DeleteRomancesTopicArn || true)
QUEUE_ARN=$(cf_output "$STACK_DATA" DeleteRomancesQueueArn || true)
QUEUE_URL=$(cf_output "$STACK_DATA" DeleteRomancesQueueUrl || true)

# 2) Store them under a neat SSM prefix
PREFIX="/${APP}/${ENV}"

[[ -n "$COUNTERS_NAME" ]] && put_param "$PREFIX/ddb/counters/name" "$COUNTERS_NAME"
[[ -n "$ROMANCES_NAME" ]]   && put_param "$PREFIX/ddb/romances/name" "$ROMANCES_NAME"
[[ -n "$TOPIC_ARN" ]]       && put_param "$PREFIX/sns/delete-romances/arn" "$TOPIC_ARN"
[[ -n "$QUEUE_ARN" ]]       && put_param "$PREFIX/sqs/delete-romances/arn" "$QUEUE_ARN"
[[ -n "$QUEUE_URL" ]]       && put_param "$PREFIX/sqs/delete-romances/url" "$QUEUE_URL"

echo "Done. Parameters saved under prefix: $PREFIX"