#!/usr/bin/env bash
# Real-world smoke test: scans pinned commits of popular open-source repositories and
# fails if any of them produces ERROR-level findings, invalid JSON/SARIF, or a crash.
# These repos contain no real secrets, so any ERROR here is a false positive.
#
# Usage: scripts/smoke-realworld.sh [path/to/archguard]
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
bin="${1:-}"
if [ -z "$bin" ]; then
  bin="$(mktemp -d)/archguard"
  go build -o "$bin" "$root/cmd/archguard"
fi
bin="$(cd "$(dirname "$bin")" && pwd)/$(basename "$bin")" # absolute: we cd into each repo
config="$root/testdata/smoke/archguard.yaml"
work="${SMOKE_WORKDIR:-$(mktemp -d)}"

# name  repository  pinned commit
repos=(
  "cobra     https://github.com/spf13/cobra.git          adbc8813901bba65827259daa8e22ff94ec1f30e"
  "gin       https://github.com/gin-gonic/gin.git        43fe48e8a0f44af783116cdb010725e6bb50255f"
  "fastapi   https://github.com/fastapi/fastapi.git      33d411dbc3236275dd64d200bfe18d5d60a49b2e"
  "commerce  https://github.com/vercel/commerce.git      3761e52e60df9c6a316e067dbfd7032e494d3634"
  "terraform https://github.com/hashicorp/terraform.git  0abcc9a9abb9519da53fb47422efec6f88dcf502"
)

failed=0
for entry in "${repos[@]}"; do
  read -r name url sha <<<"$entry"
  dir="$work/$name"
  if [ ! -d "$dir/.git" ]; then
    git init -q "$dir"
    git -C "$dir" remote add origin "$url"
  fi
  git -C "$dir" fetch -q --depth 1 origin "$sha"
  git -C "$dir" checkout -q --detach FETCH_HEAD

  set +e
  (cd "$dir" && "$bin" scan --config "$config" --format json >"$work/$name.json" 2>"$work/$name.err")
  code=$?
  (cd "$dir" && "$bin" scan --config "$config" --format sarif >"$work/$name.sarif" 2>>"$work/$name.err")
  summary="$(python3 - "$work/$name.json" "$work/$name.sarif" 2>&1 <<'PY'
import json, sys
res = json.load(open(sys.argv[1]))
sarif = json.load(open(sys.argv[2]))
assert isinstance(res["issues"], list), "issues must be a JSON array"
assert sarif["version"] == "2.1.0" and isinstance(sarif["runs"][0]["results"], list), "invalid SARIF"
errors = [i for i in res["issues"] if i["severity"] == "ERROR"]
print(f"{len(res['issues'])} issues ({len(errors)} errors) in {res['scan_time_ms']}ms")
for i in errors:
    print(f"    ERROR {i['rule_id']} {i['file_path']}:{i.get('line', 0)} {i['message']}")
PY
)"
  valid=$?
  set -e

  if [ "$code" -eq 0 ] && [ "$valid" -eq 0 ]; then
    echo "✅ $name: $summary"
  else
    echo "❌ $name: exit $code"
    if [ "$valid" -eq 0 ]; then echo "$summary"; else echo "    invalid output: $(echo "$summary" | tail -1)"; fi
    [ -s "$work/$name.err" ] && sed 's/^/    stderr: /' "$work/$name.err"
    failed=1
  fi
done

exit "$failed"
