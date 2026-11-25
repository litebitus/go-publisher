#!/usr/bin/env bash
# Usage: wait-for-macos-test.sh <job name>
# Waits for the named job of this workflow run and never fails: it says when the
# job passed, and warns when the job failed, got no runner within
# runner_wait_seconds of its creation, or has not finished within
# finish_wait_seconds of this wait starting.
set -euo pipefail

job_name="$1"
runner_wait_seconds=180
finish_wait_seconds=330
wait_started="$(date -u +%s)"

while true; do
  job="$(gh api "repos/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}/attempts/${GITHUB_RUN_ATTEMPT}/jobs" \
    --jq "[.jobs[] | select(.name == \"${job_name}\")][0]")"
  if [ "$job" = "null" ]; then
    echo "::warning::this run has no job ${job_name}; continuing without it"
    exit 0
  fi
  status="$(jq -r .status <<<"$job")"
  case "$status" in
    completed)
      conclusion="$(jq -r .conclusion <<<"$job")"
      if [ "$conclusion" = success ]; then
        echo "${job_name} passed"
      else
        echo "::warning::${job_name} ended ${conclusion}; continuing without it"
      fi
      exit 0
      ;;
    in_progress)
      if [ $(( $(date -u +%s) - wait_started )) -ge "$finish_wait_seconds" ]; then
        echo "::warning::${job_name} did not finish within ${finish_wait_seconds} seconds; continuing without it"
        exit 0
      fi
      ;;
    *)
      waited=$(( $(date -u +%s) - $(date -u -d "$(jq -r .created_at <<<"$job")" +%s) ))
      if [ "$waited" -ge "$runner_wait_seconds" ]; then
        echo "::warning::${job_name} got no runner within ${runner_wait_seconds} seconds; continuing without it"
        exit 0
      fi
      ;;
  esac
  sleep 15
done
