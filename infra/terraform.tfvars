project_name       = "contbench"
project_id         = "contbench"
service_account_id = "benchbot"
functions = [
  { name = "noop", trigger_type = "http" },
  { name = "env", trigger_type = "http" },
  { name = "watch", trigger_type = "http" },
  { name = "enqueue", trigger_type = "firestore" },
]

worker_machine_type     = "n1-standard-1"
worker_min_cpu_platform = "Intel Skylake"
