#!/bin/bash -ex

# Parameters ----------------------------------------------------------------

root=$(pwd)
infra="${root}/infra"

# Register Cleanup Function -------------------------------------------------

__cleanup ()
{
    rm -rf ${root}/fn/*/vendor ${infra}/{fn,tmp,dist.tar.gz}
}

trap __cleanup EXIT

# Build Distribution --------------------------------------------------------

GOOS=linux GOARCH=amd64 ./script/dist.sh ${infra}/dist.tar.gz

# Update Vendor Directories -------------------------------------------------

for dir in fn/*; do
    cd ${dir}
    go build
    go mod tidy
    go mod vendor
    cd -
done

# Copy functions to artifacts directory -------------------------------------

cp -r fn ${infra}
rm -f ${infra}/fn/*/go.*

# Run Terraform -------------------------------------------------------------

cd ${infra}
terraform apply
