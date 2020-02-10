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

variable "functions_path" {
  default = "../fn"
}

variable "functions" {
  type    = list(string)
  default = ["watch", "launch"]
}

variable "functions_runtime" {
  default = "go113"
}
