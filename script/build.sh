#!/usr/bin/env bash

set -e

pushd () {
    command pushd "$@" > /dev/null
}

popd () {
    command popd > /dev/null
}

SCRIPT_PATH=$( cd "$(dirname "$0")" ; pwd -P )

pushd "$SCRIPT_PATH/.."

APP=bucketblocker

# We only build for Mac OS
ARCHITECTURES=("arm64" "amd64")

for ARCH in "${ARCHITECTURES[@]}"; do
    echo ""
    echo "=== Creating release for Darwin $ARCH ==="
    GOOS=darwin GOARCH=$ARCH go build -o "$APP" main.go

    mkdir -p "release/darwin-$ARCH"
    pushd "release/darwin-$ARCH"

    TAR_NAME="${APP}_darwin-${ARCH}.tar.gz"
    tar -czf "$TAR_NAME" "../../$APP"

    SHA256_SUM=$(shasum -a 256 "$TAR_NAME" | awk '{ print $1 }')
    echo "The following is the SHA256 sum for the '$TAR_NAME' bundle:"
    echo "$SHA256_SUM"

    popd

    if [ -f "$GITHUB_ENV" ]; then
        echo "SHA256_SUM_$ARCH=$SHA256_SUM" >> "$GITHUB_ENV"
    fi
done
