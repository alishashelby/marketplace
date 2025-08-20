#!/bin/bash

exit_code=0
dirs=$(find . -type f -name "*.go" -exec dirname {} \; | sort -u)

for dir in $dirs; do
  echo "Formatting Swagger docs in: $dir"
    if ! swag fmt "$dir"; then
      exit_code=1
      break
    fi

  echo "Linting: $dir"
  if ! golangci-lint run --config=.github/.golangci.yml "$dir"; then
    exit_code=1
    break
  fi
done

exit $exit_code