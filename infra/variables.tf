variable "project_name" {
  default = "contbench"
}

variable "project_id" {
  default = "contbench"
}

variable "region" {
  default = "us-central1"
}

variable "zone" {
  default = "us-central1-a"
}

variable "functions" {
  type    = list(string)
  default = ["noop", "env", "watch"]
}

variable "functions_runtime" {
  default = "go113"
}
