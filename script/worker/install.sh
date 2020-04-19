#!/bin/bash -e

project_name="cb"
deploy_dir="/opt/${project_name}"
log_dir="/var/log/${project_name}"
tmp_dir="/var/tmp/${project_name}"
config_dir="/etc/${project_name}"

coordinator='https://us-central1-contbench.cloudfunctions.net/coordinator'
name=$(hostname)

# Install Required Packages -------------------------------------------------

apt-get update
apt-get install -y \
    build-essential \
    pkg-config \
    supervisor \
    ;

# Clean Up Previous Install -------------------------------------------------

supervisorctl stop all
rm -rf ${deploy_dir} ${log_dir} ${tmp_dir} ${config_dir}

# Install to Deploy Directory -----------------------------------------------

mkdir -p ${deploy_dir}
cp -r * ${deploy_dir}

# Configure Supervisor Processes --------------------------------------------

artifacts_dir="${tmp_dir}/artifacts"
athens_storage_dir="${tmp_dir}/athens"
mkdir -p ${config_dir} ${log_dir} ${artifacts_dir} ${athens_storage_dir}

athens_port="3000"
athens_config="${config_dir}/athens.toml"
cat > ${athens_config} <<EOF
GoBinary = "${deploy_dir}/go/bin/go"
GoEnv = "production"
GoBinaryEnvVars = ["GOPROXY=proxy.golang.org,direct"]
GoGetWorkers = 10
ProtocolWorkers = 30
CloudRuntime = "none"
Port = ":${athens_port}"
Timeout = 300
LogLevel = "info"
DownloadMode = "sync"
StorageType = "disk"
[Storage]
    [Storage.Disk]
        RootPath = "${athens_storage_dir}"
EOF
chmod 0600 ${athens_config}

cat > /etc/supervisor/conf.d/${project_name}.conf <<EOF
[group:${project_name}]
programs=worker,athens

[program:worker]
command=${deploy_dir}/bin/worker run -coordinator ${coordinator} -name ${name} -artifacts ${artifacts_dir} -goproxy http://localhost:${athens_port}
autostart=false
autorestart=false
stdout_logfile=${log_dir}/worker.out
stdout_logfile_maxbytes=64MB
stdout_logfile_backups=2
stderr_logfile=${log_dir}/worker.err
stderr_logfile_maxbytes=64MB
stderr_logfile_backups=2

[program:athens]
command=${deploy_dir}/bin/athens -config_file ${athens_config}
autostart=true
autorestart=true
stdout_logfile=${log_dir}/athens.out
stdout_logfile_maxbytes=64MB
stdout_logfile_backups=2
stderr_logfile=${log_dir}/athens.err
stderr_logfile_maxbytes=64MB
stderr_logfile_backups=2
EOF

supervisorctl reload
