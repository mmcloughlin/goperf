#!/bin/bash -e

# Setup Working Directory ---------------------------------------------------

mkdir -p ${deploy_dir}
tmpdir=$(mktemp -d)

# Install Required Packages -------------------------------------------------

apt-get update
apt-get install -y supervisor

# Download and Unpack Deploy Package ----------------------------------------

archive_name=$(basename ${dist_archive_gs_uri})
archive_path="$tmp_dir/$archive_name"

gsutil cp ${dist_archive_gs_uri} $archive_path

tar xzf $archive_path --strip-components=1 -C ${deploy_dir}
