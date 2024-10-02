#!/usr/bin/env bash

set -e

APP="fsbp-fix"

pushd () {
    command pushd "$@" > /dev/null
}

popd () {
    command popd > /dev/null
}

SCRIPT_PATH=$( cd "$(dirname "$0")" ; pwd -P )
pushd "$SCRIPT_PATH/.."



# We only build for Mac OS
ARCHITECTURES=("arm64" "amd64")

for ARCH in "${ARCHITECTURES[@]}"; do
    echo ""
    echo "=== Creating release for Darwin $ARCH ==="
    GOOS=darwin GOARCH=$ARCH go build -o "$APP" main.go

    mkdir -p "release/darwin-$ARCH"

    TAR_NAME="${APP}_darwin-${ARCH}.tar.gz"
    tar -czf "release/darwin-${ARCH}/${TAR_NAME}" "$APP"

    pushd "release/darwin-$ARCH"

    SHA256_SUM=$(shasum -a 256 "$TAR_NAME" | awk '{ print $1 }')
    echo "The following is the SHA256 sum for the '$TAR_NAME' bundle:"
    echo "$SHA256_SUM"

    popd

    ## Create build summary if running in Github Actions
    if [ -f "$GITHUB_STEP_SUMMARY" ]; then
        echo "### Darwin ${ARCH} SHA256 sum" >> $GITHUB_STEP_SUMMARY 
        echo "$SHA256_SUM" >> $GITHUB_STEP_SUMMARY 
    fi
done
