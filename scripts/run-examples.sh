#!/usr/bin/env bash
set -uo pipefail

examples=(
  "delete/byid:./cmd/delete/byid"
  "delete/query:./cmd/delete/query"
  "document/savegetdelete:./cmd/document/savegetdelete"
  "query/aggregateavg:./cmd/query/aggregateavg"
  "query/aggregateswithgrouping:./cmd/query/aggregateswithgrouping"
  "query/basic:./cmd/query/basic"
  "query/compound:./cmd/query/compound"
  "query/findbyid:./cmd/query/findbyid"
  "query/firstornull:./cmd/query/firstornull"
  "query/innerquery:./cmd/query/innerquery"
  "query/inpartition:./cmd/query/inpartition"
  "query/list:./cmd/query/list"
  "query/notinnerquery:./cmd/query/notinnerquery"
  "query/orderby:./cmd/query/orderby"
  "query/resolver:./cmd/query/resolver"
  "query/searchbyresolverfields:./cmd/query/searchbyresolverfields"
  "query/select:./cmd/query/select"
  "query/sortingandpaging:./cmd/query/sortingandpaging"
  "query/update:./cmd/query/update"
  "save/basic:./cmd/save/basic"
  "save/batchsave:./cmd/save/batchsave"
  "save/cascade:./cmd/save/cascade"
  "save/cascadebuilder:./cmd/save/cascadebuilder"
  "schema/basic:./cmd/schema/basic"
  "secrets/basic:./cmd/secrets/basic"
  "stream/close:./cmd/stream/close"
  "stream/createevents:./cmd/stream/createevents"
  "stream/deleteevents:./cmd/stream/deleteevents"
  "stream/querystream:./cmd/stream/querystream"
  "stream/updateevents:./cmd/stream/updateevents"
  "seed:./cmd/seed"
)

passed=0
failed=0
line_width=40
marker="example: completed"
green=$'\033[32m'
red=$'\033[31m'
reset=$'\033[0m'
declare -a failed_names=()
declare -a failed_logs=()

pushd "$(dirname "${BASH_SOURCE[0]}")/../examples" >/dev/null

for entry in "${examples[@]}"; do
  name=${entry%%:*}
  path=${entry#*:}
  output=$(go run "$path" 2>&1)
  status="FAIL"
  if [[ $? -eq 0 && "$output" == *"$marker"* ]]; then
    status="PASS"
    ((passed++))
  else
    ((failed++))
    failed_names+=("$name")
    failed_logs+=("$output")
  fi

  dots_count=$((line_width - ${#name} - ${#status}))
  if ((dots_count < 1)); then
    dots_count=1
  fi
  dots=$(printf '%*s' "$dots_count" '' | tr ' ' '.')
  color="$red"
  if [[ "$status" == "PASS" ]]; then
    color="$green"
  fi
  printf '%s%s%s%s%s\n' "$name" "$dots" "$color" "$status" "$reset"
done

popd >/dev/null

if ((failed > 0)); then
  echo
  echo "failed logs"
  echo "-----------"
  for i in "${!failed_names[@]}"; do
    echo "[$((i+1))] ${failed_names[$i]}"
    echo "${failed_logs[$i]}"
    echo "-----------"
  done
fi

echo
echo "totals"
echo "------"
printf 'PASSED: %s%d%s\n' "$green" "$passed" "$reset"
printf 'FAILED: %s%d%s\n' "$red" "$failed" "$reset"

if ((failed > 0)); then
  exit 1
fi
