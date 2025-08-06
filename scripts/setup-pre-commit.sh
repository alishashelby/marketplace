#!/bin/bash

pip install pre-commit
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3
pre-commit install
