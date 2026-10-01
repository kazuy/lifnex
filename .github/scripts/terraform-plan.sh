#!/usr/bin/env bash

set -euo pipefail

plan_file="${1:?Usage: terraform-plan.sh <plan-file>}"

: "${GITHUB_OUTPUT:?GITHUB_OUTPUT must be set}"
: "${GITHUB_STEP_SUMMARY:?GITHUB_STEP_SUMMARY must be set}"

set +e
terraform -chdir=infra plan -lock=false -detailed-exitcode -out="${plan_file}" -no-color
exit_code=$?
set -e

case "${exit_code}" in
  0)
    echo "has_changes=false" >> "${GITHUB_OUTPUT}"
    ;;
  2)
    echo "has_changes=true" >> "${GITHUB_OUTPUT}"
    ;;
  *)
    exit "${exit_code}"
    ;;
esac

plan_json="$(terraform -chdir=infra show -json "${plan_file}")"
add_count="$(jq '[.resource_changes[]? | select(.change.actions | index("create"))] | length' <<< "${plan_json}")"
change_count="$(jq '[.resource_changes[]? | select(.change.actions == ["update"])] | length' <<< "${plan_json}")"
destroy_count="$(jq '[.resource_changes[]? | select(.change.actions | index("delete"))] | length' <<< "${plan_json}")"

{
  echo "## Terraform Plan"
  echo
  echo "| Action | Count |"
  echo "| --- | ---: |"
  echo "| Add | ${add_count} |"
  echo "| Change | ${change_count} |"
  echo "| Destroy | ${destroy_count} |"
} >> "${GITHUB_STEP_SUMMARY}"
