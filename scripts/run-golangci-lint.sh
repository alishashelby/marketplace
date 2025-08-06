#!/bin/bash

exit_code=0
dirs=$(find . -type f -name "*.go" -exec dirname {} \; | sort -u)

for dir in $dirs; do
  echo "Linting: $dir"
  if ! golangci-lint run --config=.github/.golangci.yml "$dir"; then
    exit_code=1
  fi
done

exit $exit_code