project_name       = "contbench"
project_id         = "contbench"
service_account_id = "benchbot"
functions = [
  { name = "noop", trigger_type = "http" },
  { name = "env", trigger_type = "http" },
  { name = "watch", trigger_type = "http" },
  { name = "ingest", trigger_type = "result" },
  { name = "coordinator", trigger_type = "http" },
  { name = "staletimeout", trigger_type = "http" },
  { name = "dashboard", trigger_type = "http" },
]

worker_machine_type     = "n1-standard-2"
worker_min_cpu_platform = "Intel Skylake"
