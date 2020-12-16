terraform {
  required_version = "~> 0.13"
  required_providers {
    aws  = "~> 3"
    null = "~> 2.1"
  }
}
provider "aws" {
  region = var.region
}
