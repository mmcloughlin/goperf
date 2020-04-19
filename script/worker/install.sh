#!/bin/bash -e

project_name="cb"
deploy_dir="/opt/${project_name}"
log_dir="/var/log/${project_name}"
tmp_dir="/var/tmp/${project_name}"

coordinator='https://us-central1-contbench.cloudfunctions.net/coordinator'
name=$(hostname)

# Install Required Packages -------------------------------------------------

apt-get update
apt-get install -y \
    build-essential \
    pkg-config \
    supervisor \
    ;

# Install to Deploy Directory -----------------------------------------------

rm -rf ${deploy_dir}
install -D -t ${deploy_dir}/bin bin/*

# Configure Supervisor Processes --------------------------------------------

artifacts_dir="${tmp_dir}/artifacts"
mkdir -p ${log_dir} ${artifacts_dir}

cat > /etc/supervisor/conf.d/${project_name}.conf <<EOF
[program:${project_name}-worker]
command=${deploy_dir}/bin/worker run -coordinator ${coordinator} -name ${name} -artifacts ${artifacts_dir}
process_name=${project_name}-worker
autostart=false
autorestart=false
stdout_logfile=${log_dir}/worker.out
stdout_logfile_maxbytes=64MB
stdout_logfile_backups=2
stderr_logfile=${log_dir}/worker.err
stderr_logfile_maxbytes=64MB
stderr_logfile_backups=2
EOF

supervisorctl reload
