# If vpc_id is not provided - use default VPC's as a resource
resource "aws_default_vpc" "default" {
  tags = {
    Name = "Default VPC"
  }
}

resource "aws_default_subnet" "default_az0" {
  availability_zone = var.availability_zones[0]
  tags = {
    Name                      = "Default subnet for cluster.dev in AZ1"
    "cluster.dev/subnet_type" = "default2"
  }
}

resource "aws_default_subnet" "default_az1" {
  availability_zone = var.availability_zones[1]
  tags = {
    Name                      = "Default subnet for cluster.dev in AZ2"
    "cluster.dev/subnet_type" = "default2"
  }
}
