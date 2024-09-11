#!/usr/bin/env bash

TOOL_NAME="bucket-blocker"
INCORRECT_NAME="bucketblocker"

CURRENT_FILE=$(basename "${BASH_SOURCE}")
ROOT_DIR=$(git rev-parse --show-toplevel)
BAD_NAMES=$(grep -ril "$INCORRECT_NAME" "$ROOT_DIR" --exclude "$CURRENT_FILE")

if [ -n "$BAD_NAMES" ]; then
    echo "$TOOL_NAME has a dash in it, and inconsistencies in the codebase can cause major problems, please correct the spelling." >&2
    printf "The following files contain the incorrect string '%s':\n\n" "$INCORRECT_NAME" >&2
    echo "$BAD_NAMES" >&2
    exit 1;
else
    echo "No files contain the name '$INCORRECT_NAME'."
    exit 0;
fi
