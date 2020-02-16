#!/bin/bash -ex

archive=$1

# Parameters ----------------------------------------------------------------

name="cb"
version=$(git describe --always --dirty)
export GOARCH=${GOARCH:-amd64}
export GOOS=${GOOS:-$(go env GOOS)}

# Prepare Workspace ---------------------------------------------------------

workdir=$(mktemp -d)
pkgdir="${workdir}/${name}"
bindir="${pkgdir}/bin"

mkdir -p ${workdir} ${pkgdir} ${bindir}

# Build ---------------------------------------------------------------------

GOBIN=${bindir} go install ./app/cmd/worker

# Versioning ----------------------------------------------------------------

echo ${version} > ${pkgdir}/VERSION

# Build Archive -------------------------------------------------------------

tar -C ${workdir} -czf ${archive} ${name}

# Cleanup Workspace ---------------------------------------------------------

rm -rf ${workdir}
