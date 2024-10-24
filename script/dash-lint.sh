#!/usr/bin/env bash

TOOL_NAME=$1
ROOT_DIR=$(git rev-parse --show-toplevel)

function bad_delimiter_check() {
    DELIMITER=$1
    # shellcheck disable=SC2116 # Without echo, WRONG_NAME will be empty
    WRONG_NAME=$(echo "${TOOL_NAME//-/$DELIMITER}")
    echo "Checking for $WRONG_NAME in the codebase..."
    FAILURES=$(grep -ril "$WRONG_NAME" "$ROOT_DIR")

    if [ -n "$FAILURES" ]; then
        echo "$TOOL_NAME has a dash in it, and inconsistencies in the codebase can cause major problems, please correct the spelling." >&2
        printf "The following files contain the incorrect string '%s':\n\n" "$WRONG_NAME" >&2
        echo "$FAILURES" >&2
        exit 1;
    else
        printf "No files contain the erroneous name: '%s'.\n\n" "$WRONG_NAME"
    fi
}

bad_delimiter_check "_"
bad_delimiter_check " "
bad_delimiter_check ""
