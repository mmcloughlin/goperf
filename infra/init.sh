#!/bin/bash -e

# Setup Working Directory ---------------------------------------------------

mkdir -p ${deploy_dir}
tmpdir=$(mktemp -d)

# Install Required Packages -------------------------------------------------

apt-get update
apt-get install -y \
    build-essential \
    pkg-config \
    supervisor \
    ;

# Download and Unpack Deploy Package ----------------------------------------

archive_name=$(basename ${dist_archive_gs_uri})
archive_path="$tmp_dir/$archive_name"

gsutil cp ${dist_archive_gs_uri} $archive_path

tar xzf $archive_path --strip-components=1 -C ${deploy_dir}

# Configure Supervisor Processes --------------------------------------------

mkdir -p ${log_dir}
cat > /etc/supervisor/conf.d/${project_name}.conf <<EOF
[program:${project_name}-worker]
command=${deploy_dir}/bin/worker run
process_name=${project_name}-worker
autostart=true
autorestart=true
stdout_logfile=${log_dir}/worker.out
stdout_logfile_maxbytes=8MB
stdout_logfile_backups=2
stderr_logfile=${log_dir}/worker.err
stderr_logfile_maxbytes=8MB
stderr_logfile_backups=2
EOF

supervisorctl reload
