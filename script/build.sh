#!/bin/bash -ex

# Build Main Module ---------------------------------------------------------

go build ./...

# Build Cloud Function Modules ----------------------------------------------

for dir in fn/*; do
    cd ${dir}
    go build
    go mod tidy
    cd -
done
