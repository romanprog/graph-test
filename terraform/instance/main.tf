data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

data "aws_subnet" "instance_subnet" {
  id = var.subnet_id
}

resource "aws_security_group" "instance" {
  vpc_id = data.aws_subnet.instance_subnet.vpc_id
  name   = var.cluster_name
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "foo" {
  ami           = data.aws_ami.ubuntu.id
  associate_public_ip_address = true
  instance_type = var.instance_type

  subnet_id = var.subnet_id

  vpc_security_group_ids = [
    aws_security_group.instance.id,
  ]
}