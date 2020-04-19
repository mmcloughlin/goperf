#!/bin/bash -ex

archive=$1

# Parameters ----------------------------------------------------------------

name="cb"
pkg="github.com/mmcloughlin/${name}"
version=$(git describe --always --dirty)
gitsha=$(git rev-parse HEAD)
scriptdir=$(dirname $0)

athens_version="v0.8.1"
go_version="1.14.2"

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

# Build Athens Proxy --------------------------------------------------------

mkdir ${workdir}/athens
cd ${workdir}/athens
wget -O athens.tar.gz "https://github.com/gomods/athens/archive/${athens_version}.tar.gz"
tar xf athens.tar.gz --strip-components=1
make build-ver VERSION=${athens_version}
mv athens ${bindir}
cd -

# Package Go ----------------------------------------------------------------

go_archive="${workdir}/go.tar.gz"
wget -O ${go_archive} https://dl.google.com/go/go${go_version}.${GOOS}-${GOARCH}.tar.gz
tar -C ${pkgdir} -xzf ${go_archive}

# Versioning ----------------------------------------------------------------

echo ${version} > ${pkgdir}/VERSION

# Install Script ------------------------------------------------------------

cp ${scriptdir}/install.sh ${pkgdir}

# Build Archive -------------------------------------------------------------

tar --sort=name --owner=root:0 --group=root:0 --mtime='UTC 2020-01-01' \
    -C ${workdir} -c ${name} | gzip -n > ${archive}

# Cleanup Workspace ---------------------------------------------------------

rm -rf ${workdir}
