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

# Try multiple keys (backward-compatible); return first non-empty
cf_output_any() {
  local stack="$1"; shift
  local val
  for key in "$@"; do
    val="$(cf_output "$stack" "$key" 2>/dev/null || true)"
    if [[ -n "$val" && "$val" != "None" ]]; then
      echo "$val"
      return 0
    fi
  done
  echo ""
}

# 1) Read outputs from DataStack (primary stream)
COUNTERS_NAME="$(cf_output_any "$STACK_DATA" CountersTableName)"
ROMANCES_NAME="$(cf_output_any "$STACK_DATA" RomancesTableName)"

TOPIC_ARN="$(cf_output_any "$STACK_DATA" DeleteRomancesFifoTopicArn DeleteRomancesTopicArn)"
QUEUE_ARN="$(cf_output_any "$STACK_DATA" DeleteRomancesFifoQueueArn DeleteRomancesQueueArn)"
QUEUE_URL="$(cf_output_any "$STACK_DATA" DeleteRomancesFifoQueueUrl DeleteRomancesQueueUrl)"

# 1b) Read outputs from DataStack (group stream)
GROUP_TOPIC_ARN="$(cf_output_any "$STACK_DATA" DeleteRomancesGroupFifoTopicArn DeleteRomancesGroupTopicArn)"
GROUP_QUEUE_ARN="$(cf_output_any "$STACK_DATA" DeleteRomancesGroupFifoQueueArn DeleteRomancesGroupQueueArn)"
GROUP_QUEUE_URL="$(cf_output_any "$STACK_DATA" DeleteRomancesGroupFifoQueueUrl DeleteRomancesGroupQueueUrl)"

# 2) Store them under a neat SSM prefix
PREFIX="/${APP}/${ENV}"

[[ -n "$COUNTERS_NAME" ]]      && put_param "$PREFIX/ddb/counters/name" "$COUNTERS_NAME"
[[ -n "$ROMANCES_NAME" ]]      && put_param "$PREFIX/ddb/romances/name" "$ROMANCES_NAME"

[[ -n "$TOPIC_ARN" ]]          && put_param "$PREFIX/sns/delete-romances/arn" "$TOPIC_ARN"
[[ -n "$QUEUE_ARN" ]]          && put_param "$PREFIX/sqs/delete-romances/arn" "$QUEUE_ARN"
[[ -n "$QUEUE_URL" ]]          && put_param "$PREFIX/sqs/delete-romances/url" "$QUEUE_URL"

# Group stream
[[ -n "$GROUP_TOPIC_ARN" ]]    && put_param "$PREFIX/sns/delete-romances-group/arn" "$GROUP_TOPIC_ARN"
[[ -n "$GROUP_QUEUE_ARN" ]]    && put_param "$PREFIX/sqs/delete-romances-group/arn" "$GROUP_QUEUE_ARN"
[[ -n "$GROUP_QUEUE_URL" ]]    && put_param "$PREFIX/sqs/delete-romances-group/url" "$GROUP_QUEUE_URL"

echo "Done. Parameters saved under prefix: $PREFIX"
