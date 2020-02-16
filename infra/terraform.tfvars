project_name = "contbench"
project_id   = "contbench"
functions = [
  { name = "noop", trigger_type = "http" },
  { name = "env", trigger_type = "http" },
  { name = "watch", trigger_type = "http" },
  { name = "enqueue", trigger_type = "firestore" },
]
dist_path = "dist.tar.gz"
