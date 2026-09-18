#!/usr/bin/env bash

set -e

version="${1:-${GITHUB_REF_NAME#v}}"
if [ -z "$version" ]; then
  version=$(git describe --tags --always)
fi

rm -rf ./dist
mkdir -p ./dist

targets=(
  "x86_64-unknown-linux-gnu:linux:amd64"
  "aarch64-unknown-linux-gnu:linux:arm64"
  "x86_64-apple-darwin:darwin:amd64"
  "aarch64-apple-darwin:darwin:arm64"
)

for entry in "${targets[@]}"; do
  IFS=: read -r target goos goarch <<<"$entry"

  binary_file="gira-$version-$target"
  echo "Compiling for $target..."

  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build \
    -trimpath \
    -ldflags "-s -w -X main.version=$version" \
    -o "./dist/$binary_file" \
    .

  echo "Creating tarball for $target..."
  tar -czvf "./dist/$binary_file.tar.gz" -C ./dist "$binary_file"

  echo "Removing binary file $binary_file"
  rm "./dist/$binary_file"
done

echo "Build complete."
