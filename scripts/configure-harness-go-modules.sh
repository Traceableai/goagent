#!/usr/bin/env bash

set -euo pipefail

if [[ ! -f go.mod ]]; then
  echo "go.mod file not found in current directory" >&2
  exit 1
fi

if [[ -z "${HARNESS_CODE_READ_TOKEN:-}" ]]; then
  echo "HARNESS_CODE_READ_TOKEN is required" >&2
  exit 1
fi

HARNESS_CODE_BASE_URL="https://git:${HARNESS_CODE_READ_TOKEN}@git0.harness.io/l7B_kbSEQD2wjrM7PShm5w/PROD/Traceable"

PUBLIC_REPOS_REGEX='^(goagent|goagent-src|goagent-example|agent-config)$'

CURRENT_REPO="$(basename "$(pwd)")"

declare -A seen_repos=()
declare -A local_replace_repos=()

while read -r repo; do
  [[ -n "${repo}" ]] && local_replace_repos["$repo"]=1
done < <(
  awk '
    /=>[[:space:]]*\.\.?\// {
      if (match($0, /github.com\/Traceableai\/[A-Za-z0-9._\/-]+/)) {
        mod = substr($0, RSTART, RLENGTH)
        sub(/^github.com\/Traceableai\//, "", mod)
        split(mod, parts, "/")
        print parts[1]
      }
    }
  ' go.mod | sort -u
)

while read -r module; do
  repo_path="${module#github.com/Traceableai/}"
  repo="${repo_path%%/*}"
  [[ -z "${repo}" ]] && continue

  if [[ "${repo}" =~ ${PUBLIC_REPOS_REGEX} ]]; then
    continue
  fi

  if [[ "${repo}" == "${CURRENT_REPO}" ]] || [[ -n "${local_replace_repos[$repo]:-}" ]]; then
    continue
  fi

  [[ -n "${seen_repos[$repo]:-}" ]] && continue
  seen_repos["$repo"]=1

  git config --global \
    url."${HARNESS_CODE_BASE_URL}/${repo}.git".insteadOf \
    "https://github.com/Traceableai/${repo}"
  echo "Configured Harness Code rewrite for Traceableai/${repo}"
done < <(
  awk '{
    while (match($0, /github.com\/Traceableai\/[A-Za-z0-9._\/-]+/)) {
      print substr($0, RSTART, RLENGTH)
      $0 = substr($0, RSTART + RLENGTH)
    }
  }' go.mod | sort -u
)

go env -w GOPRIVATE=github.com/Traceableai/*
go env -w GONOSUMDB=github.com/Traceableai/*
go env -w GOPROXY=direct

echo "Configured Go private module environment for Traceableai modules"
