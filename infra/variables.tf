variable "project_name" {
  default = "contbench"
}

variable "project" {
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
  default = ["watch"]
}

variable "functions_runtime" {
  default = "go111"
}
