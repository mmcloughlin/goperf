#!/bin/bash -ex

# Maintenance Tasks ---------------------------------------------------------

./script/fmt.sh
./script/generate.sh

# Build Main Module ---------------------------------------------------------

go build ./...

# Build Cloud Function Modules ----------------------------------------------

for dir in fn/*; do
    cd ${dir}
    go build
    cd -
done
