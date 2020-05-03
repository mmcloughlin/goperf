project_name       = "contbench"
project_id         = "contbench"
service_account_id = "benchbot"
functions = {
  noop         = { trigger_type = "http" },
  env          = { trigger_type = "http" },
  watch        = { trigger_type = "http" },
  ingest       = { trigger_type = "result", memory = 512, timeout = 480 },
  changedetect = { trigger_type = "http", memory = 2048, timeout = 480 },
  coordinator  = { trigger_type = "http", memory = 256 },
  staletimeout = { trigger_type = "http" },
  dashboard    = { trigger_type = "http", memory = 512 },
}
