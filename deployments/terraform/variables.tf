variable "image" {
  default = "Ubuntu 20.04 LTS Focal Fossa"
}

variable "flavor" {
  default = "a4-ram8-disk80-perf1"
}

variable "ssh_key_file" {
  default = "~/.ssh/id_rsa.kubespray"
}

variable "ssh_user_name" {
  default = "ubuntu"
}

variable "pool" {
  default = "ext-floating1"
}

variable "volume_size" {
  type    = number
  default = 1
}