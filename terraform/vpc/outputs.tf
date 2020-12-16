output "vpc_id" {
  value = aws_default_vpc.default.id
}

output "private_subnets" {
  value = [aws_default_subnet.default_az0.id, aws_default_subnet.default_az1.id]
}

output "public_subnets" {
  value = [aws_default_subnet.default_az0.id, aws_default_subnet.default_az1.id]
}

output "vpc_cidr" {
  value = aws_default_vpc.default.cidr_block
}
