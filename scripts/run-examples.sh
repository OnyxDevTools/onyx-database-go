#!/usr/bin/env bash
set -uo pipefail

examples=(
  "delete/byid:./examples/delete/cmd/byid"
  "delete/query:./examples/delete/cmd/query"
  "document/savegetdelete:./examples/document/cmd/savegetdelete"
  "query/aggregateavg:./examples/query/cmd/aggregateavg"
  "query/aggregateswithgrouping:./examples/query/cmd/aggregateswithgrouping"
  "query/basic:./examples/query/cmd/basic"
  "query/compound:./examples/query/cmd/compound"
  "query/findbyid:./examples/query/cmd/findbyid"
  "query/firstornull:./examples/query/cmd/firstornull"
  "query/innerquery:./examples/query/cmd/innerquery"
  "query/inpartition:./examples/query/cmd/inpartition"
  "query/list:./examples/query/cmd/list"
  "query/notinnerquery:./examples/query/cmd/notinnerquery"
  "query/orderby:./examples/query/cmd/orderby"
  "query/resolver:./examples/query/cmd/resolver"
  "query/searchbyresolverfields:./examples/query/cmd/searchbyresolverfields"
  "query/select:./examples/query/cmd/select"
  "query/sortingandpaging:./examples/query/cmd/sortingandpaging"
  "query/update:./examples/query/cmd/update"
  "save/basic:./examples/save/cmd/basic"
  "save/batchsave:./examples/save/cmd/batchsave"
  "save/cascade:./examples/save/cmd/cascade"
  "save/cascadebuilder:./examples/save/cmd/cascadebuilder"
  "schema/basic:./examples/schema/cmd/basic"
  "secrets/basic:./examples/secrets/cmd/basic"
  "stream/close:./examples/stream/cmd/close"
  "stream/createevents:./examples/stream/cmd/createevents"
  "stream/deleteevents:./examples/stream/cmd/deleteevents"
  "stream/querystream:./examples/stream/cmd/querystream"
  "stream/updateevents:./examples/stream/cmd/updateevents"
  "seed:./examples/cmd/seed"
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
