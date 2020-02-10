#!/bin/bash -ex

# update vendor directory
for dir in fn/*; do
    cd ${dir}
    go build
    go mod tidy
    go mod vendor
    cd -
done

# copy functions into infra directory
rm -rf infra/fn
cp -r fn infra
rm -f infra/fn/*/go.*

# run terraform
cd infra
terraform apply
