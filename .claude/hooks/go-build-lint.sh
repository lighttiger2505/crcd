#!/usr/bin/env bash
# Claude Code PostToolUse hook:
#   .go を編集したら build + golangci-lint を実行し、問題のみ通知する（非ブロック）
set -uo pipefail

input=$(cat)
file_path=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')

# Go ソースの編集だけに反応
case "$file_path" in
  *.go) ;;
  *) exit 0 ;;
esac

cd "${CLAUDE_PROJECT_DIR:-.}" || exit 0

report=""

build_out=$(make build 2>&1)
if [ $? -ne 0 ]; then
  report+=$'### build failed\n```\n'"$build_out"$'\n```\n'
fi

lint_out=$(golangci-lint run ./... 2>&1)
if [ $? -ne 0 ]; then
  report+=$'### golangci-lint issues\n```\n'"$lint_out"$'\n```\n'
fi

# 問題なし → 静かに終了
[ -z "$report" ] && exit 0

# 非ブロックで Claude / ユーザーに通知
jq -n --arg ctx "$report" \
  '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":$ctx}}'
exit 0
