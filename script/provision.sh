#!/bin/bash -ex

# Parameters ----------------------------------------------------------------

infra="infra"
distfile="dist.tar.gz"

# Update Vendor Directories -------------------------------------------------

for dir in fn/*; do
    cd ${dir}
    go build
    go mod tidy
    go mod vendor
    cd -
done

# Copy functions to infra directory -----------------------------------------

rm -rf ${infra}/fn
cp -r fn ${infra}
rm -f ${infra}/fn/*/go.*

# Build Distribution --------------------------------------------------------

GOOS=linux GOARCH=amd64 ./script/dist.sh ${infra}/${distfile}

# Run Terraform -------------------------------------------------------------

cd ${infra}
terraform apply -var "dist_path=${distfile}"
