variable "region" {
  type        = string
  description = "The AWS region."
}

variable "instance_type" {
  type        = string
}

variable "subnet_id" {
  type        = string
}

variable "cluster_name" {
  type        = string
  description = "Name of the cluster."
}
