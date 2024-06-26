#!/usr/bin/env bash

CURRENT_FILE=$(basename "${BASH_SOURCE}")
ROOT_DIR=$(git rev-parse --show-toplevel)
BAD_NAMES=$(grep -ril "bucketblocker" "$ROOT_DIR" --exclude "$CURRENT_FILE")

if [ -n "$BAD_NAMES" ]; then
    printf "The following files contain the incorrect string 'bucketblocker':\n\n"
    echo "$BAD_NAMES"
    printf "\nbucket-blocker has a dash in it, and inconsistencies in the codebase can cause major problems, please correct the spelling.\n"
    exit 1;
else
    echo "No files contain the name 'bucketblocker'"
    exit 0;
fi
