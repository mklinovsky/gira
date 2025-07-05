#!/usr/bin/env bash

set -e

version=$(jq -r '.version' ./deno.json)

rm -rf ./dist
mkdir -p ./dist

targets=(
  "x86_64-unknown-linux-gnu"
  "aarch64-unknown-linux-gnu"
  "x86_64-apple-darwin"
  "aarch64-apple-darwin"
)

for target in "${targets[@]}"; do
  echo "Compiling for $target..."
  binary_file="dist/gira-$version-$target"

  deno compile --allow-env --allow-sys --allow-read --allow-net --output "$binary_file" --target "$target" src/main.ts

  echo "Creating tarball for $target..."
  tar -czvf "$binary_file.tar.gz" "$binary_file"

  echo "Removing binary file $binary_file"
  rm $binary_file
done

echo "Build complete."
