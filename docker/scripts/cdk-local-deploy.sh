#!/usr/bin/env sh
# Run inside the infra container to synth & deploy the DataStack to LocalStack without CDK bootstrap.

set -eux
trap 'echo "[infra] EXIT status=$?"' EXIT

# ---- Env (safe defaults) -----------------------------------------------------
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-dummy}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-dummy}"
export AWS_SESSION_TOKEN="${AWS_SESSION_TOKEN:-}"
export AWS_REGION="${AWS_REGION:-us-east-2}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-$AWS_REGION}"

export CDK_DEFAULT_ACCOUNT="${CDK_DEFAULT_ACCOUNT:-000000000000}"
export CDK_DEFAULT_REGION="${CDK_DEFAULT_REGION:-$AWS_REGION}"
export ENV_TYPE="${ENV_TYPE:-local}"

# Point SDKs to LocalStack
export AWS_ENDPOINT_URL="${AWS_ENDPOINT_URL:-http://localstack:4566}"
export AWS_ENDPOINT_URL_S3="${AWS_ENDPOINT_URL_S3:-http://localstack:4566}"
export AWS_S3_FORCE_PATH_STYLE="${AWS_S3_FORCE_PATH_STYLE:-true}"
export AWS_EC2_METADATA_DISABLED="${AWS_EC2_METADATA_DISABLED:-true}"

# Toolchain
export PATH="/usr/local/go/bin:$PATH"
export GOTOOLCHAIN="${GOTOOLCHAIN:-auto}"

echo "[infra] ENV SUMMARY: ACCOUNT=$CDK_DEFAULT_ACCOUNT REGION=$AWS_REGION ENDPOINT=$AWS_ENDPOINT_URL"

# ---- Install tools -----------------------------------------------------------
apk add --no-cache nodejs npm jq curl git aws-cli >/dev/null || true
npm config set fund false audit false progress=false >/dev/null
npm i -g aws-cdk@2 aws-cdk-local >/dev/null

echo "[infra] versions"
which go && go version
node --version
npm --version
cdklocal --version
aws --version || true
jq --version

# ---- Project layout ----------------------------------------------------------
cd /work/infra
sed -n '1,80p' cdk.json || true

# ---- Ensure LocalStack is up (one-shot log of health) -----------------------
echo "[infra] LocalStack health (one shot):"
set +e
curl -sf "$AWS_ENDPOINT_URL/_localstack/health" | jq . || true
set -e
echo "[infra] Note: follow LocalStack logs with: docker compose logs -f localstack"

# ---- Pre-clean (avoid leftovers causing rollbacks) ---------------------------
AWS="aws --endpoint-url $AWS_ENDPOINT_URL --region $AWS_REGION"
echo "[infra] pre-clean: delete old CloudFormation stack if present"
$AWS cloudformation describe-stacks --stack-name DataStack >/dev/null 2>&1 || true
# If you want to always start clean, uncomment:
# $AWS cloudformation delete-stack --stack-name DataStack || true
# $AWS cloudformation wait stack-delete-complete --stack-name DataStack || true

echo "[infra] pre-clean: drop leftover DynamoDB tables if present"
$AWS dynamodb describe-table --table-name Counters >/dev/null 2>&1 && $AWS dynamodb delete-table --table-name Counters || true
$AWS dynamodb describe-table --table-name Romances >/dev/null 2>&1 && $AWS dynamodb delete-table --table-name Romances || true

# ---- cdk list (sanity) -------------------------------------------------------
echo "[infra] cdk list"
cdklocal list || true

# ---- Synthesize & verify JSON ------------------------------------------------
echo "[infra] synth DataStack"
cdklocal context --clear
ENV_TYPE="$ENV_TYPE" cdklocal synth -j DataStack 1>/work/infra/tmp.cfn.json 2>/work/infra/synth.stderr.log || true

if ! jq -e '.Resources' /work/infra/tmp.cfn.json >/dev/null 2>&1; then
  echo "[infra] ERROR: synthesized template is not valid JSON"
  head -n 120 /work/infra/synth.stderr.log || true
  head -n 120 /work/infra/tmp.cfn.json || true
  exit 1
fi

echo "[infra] resource types in template:"
jq -r '.Resources | to_entries[] | .value.Type' /work/infra/tmp.cfn.json | sort -u

# ---- Deploy (NO LOOKUPS, NO BOOTSTRAP) --------------------------------------
echo "[infra] DEPLOY DataStack (no lookups, no bootstrap)"
ENV_TYPE="$ENV_TYPE" \
cdklocal deploy DataStack \
  --require-approval never \
  --progress events \
  --verbose \
  --no-lookups

# ---- Verify against LocalStack ----------------------------------------------
$AWS dynamodb list-tables || true
$AWS sns list-topics || true
$AWS sqs list-queues || true

echo "[infra] DONE"
