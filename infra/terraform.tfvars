project_name       = "contbench"
project_id         = "contbench"
service_account_id = "benchbot"
functions = {
  noop         = { trigger_type = "http" },
  env          = { trigger_type = "http" },
  watch        = { trigger_type = "http" },
  ingest       = { trigger_type = "result", memory = 256 },
  changedetect = { trigger_type = "http", memory = 2048, timeout = 480 },
  coordinator  = { trigger_type = "http" },
  staletimeout = { trigger_type = "http" },
  dashboard    = { trigger_type = "http", memory = 512 },
}

worker_machine_type     = "n1-standard-2"
worker_min_cpu_platform = "Intel Skylake"
