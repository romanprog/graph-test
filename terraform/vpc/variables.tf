variable "region" {
  type        = string
  description = "The AWS region."
}

variable "availability_zones" {
  type        = list
  description = "The AWS Availability Zone(s) inside region."
}

variable "cluster_name" {
  type        = string
  description = "Name of the cluster."
}
