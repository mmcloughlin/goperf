#!/bin/bash -ex

name=$1

# parameters
dir="fn/${name}"
repo="github.com/mmcloughlin/goperf"
mod="${repo}/${dir}"

# setup directory
mkdir -p ${dir}
cd ${dir}

# setup module
if [ ! -e go.mod ]; then
    go mod init ${mod}
fi

go mod edit -go=1.13
go mod edit -require=${repo}@v0.0.0
go mod edit -replace=${repo}=../..
