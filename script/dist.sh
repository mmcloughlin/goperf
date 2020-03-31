#!/bin/bash -ex

archive=$1

# Parameters ----------------------------------------------------------------

name="cb"
pkg="github.com/mmcloughlin/${name}"
version=$(git describe --always --dirty)
gitsha=$(git rev-parse HEAD)

export GOARCH=${GOARCH:-amd64}
export GOOS=${GOOS:-$(go env GOOS)}

# Prepare Workspace ---------------------------------------------------------

workdir=$(mktemp -d)
pkgdir="${workdir}/${name}"
bindir="${pkgdir}/bin"

mkdir -p ${workdir} ${pkgdir} ${bindir}

# Build ---------------------------------------------------------------------

ldflags="-X ${pkg}/meta.Version=${version} -X ${pkg}/meta.GitSHA=${gitsha}"
go build -ldflags "${ldflags}" -trimpath -o ${bindir}/worker ./app/cmd/worker

# Versioning ----------------------------------------------------------------

echo ${version} > ${pkgdir}/VERSION

# Build Archive -------------------------------------------------------------

tar --sort=name --owner=root:0 --group=root:0 --mtime='UTC 2020-01-01' \
    -C ${workdir} -c ${name} | gzip -n > ${archive}

# Cleanup Workspace ---------------------------------------------------------

rm -rf ${workdir}
