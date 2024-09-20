#!/bin/bash
# shellcheck disable=SC2164
cd "$(dirname "$0")"

go mod tidy
go mod vendor
