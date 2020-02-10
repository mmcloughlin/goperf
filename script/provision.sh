#!/bin/bash -ex

# copy functions into infra directory
rm -rf infra/fn
cp -r fn infra
rm -f infra/fn/*/go.*

# run terraform
cd infra
terraform apply
