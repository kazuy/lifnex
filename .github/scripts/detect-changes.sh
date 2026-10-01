#!/usr/bin/env bash

set -euo pipefail

mode="${1:?Usage: detect-changes.sh <pr-checks|deploy> <event-name> [base-sha] [head-sha]}"
event_name="${2:?Usage: detect-changes.sh <pr-checks|deploy> <event-name> [base-sha] [head-sha]}"
base_sha="${3:-}"
head_sha="${4:-}"

: "${GITHUB_OUTPUT:?GITHUB_OUTPUT must be set}"

app=false
infra=false

case "${mode}" in
  pr-checks)
    if [[ "${event_name}" == "workflow_dispatch" ]]; then
      app=true
      infra=true
    fi
    ;;
  deploy)
    echo "sha=$(git rev-parse HEAD)" >> "${GITHUB_OUTPUT}"

    if [[ "${event_name}" == "workflow_dispatch" ]]; then
      app=true
    fi
    ;;
  *)
    echo "Unsupported mode: ${mode}" >&2
    exit 1
    ;;
esac

if [[ "${event_name}" != "workflow_dispatch" ]]; then
  : "${base_sha:?base SHA is required}"
  : "${head_sha:?head SHA is required}"

  if [[ "${base_sha}" =~ ^0+$ ]]; then
    changed_files_command=(git diff-tree --root --no-commit-id --name-only -r "${head_sha}")
  else
    changed_files_command=(git diff --name-only "${base_sha}" "${head_sha}")
  fi

  while IFS= read -r file; do
    case "${file}" in
      app/* | Makefile)
        app=true
        ;;
      infra/*)
        infra=true
        ;;
      .github/scripts/* | .github/workflows/pr-checks.yml)
        if [[ "${mode}" == "pr-checks" ]]; then
          app=true
          infra=true
        fi
        ;;
    esac
  done < <("${changed_files_command[@]}")
fi

echo "app=${app}" >> "${GITHUB_OUTPUT}"
echo "infra=${infra}" >> "${GITHUB_OUTPUT}"
