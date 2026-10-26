#!/bin/sh
#
# Run a command in CI. If it fails, report the tail of its output as a
# GitHub Actions error annotation so the failure is visible in the run
# summary without opening the full log.
#
# Usage: script/ci-run.sh <title> <command> [args...]

set -u

TITLE="$1"
shift

LOG="$(mktemp)"
"$@" > "$LOG" 2>&1
STATUS=$?
cat "$LOG"

if [ "$STATUS" -ne 0 ]; then
	# Annotation messages are newline-escaped with %0A and are truncated
	# after a few KB, so report error lines first and then the tail.
	{
		grep -iE 'error|fatal|undefined|cannot' "$LOG" | head -n 25
		echo "----- last lines -----"
		tail -n 15 "$LOG"
	} | cut -c 1-300 | sed -e 's/%/%25/g' -e 's/\r//g' |
		awk 'BEGIN { ORS="%0A" } { print }' > "$LOG.msg"
	echo "::error title=$TITLE::$(cat "$LOG.msg")"
	rm -f "$LOG.msg"
fi

rm -f "$LOG"
exit "$STATUS"
