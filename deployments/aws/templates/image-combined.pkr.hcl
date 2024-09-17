packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

variable "deployment_name" {
  type = string
}

variable "ami_name" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "region" {
  type = string
  default = "eu-west-1"
}

variable "instance_type" {
  type = string
  default = "t2.micro"
}

variable "subnet_id" {
  type = string
}

variable "deb" {
  type = string
}

locals {
    dest_deb = "/tmp/${basename(var.deb)}"
}

source "amazon-ebs" "ubuntu" {
  ami_name      = "${var.ami_name}"
  instance_type = "${var.instance_type}"
  region        = "${var.region}"
  vpc_id        = "${var.vpc_id}"
  subnet_id     = "${var.subnet_id}"
  associate_public_ip_address = true
  tags  = {
    veraison-deployment = "${var.deployment_name}"
  }
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
      architecture        = "x86_64"
    }
    owners      = ["099720109477"]  # amazon
    most_recent = true
  }
  security_group_filter {
    filters = {
      "tag:Class": "packer"
    }
  }
  ssh_username = "ubuntu"
}

build {
  name = "veraison-combined"
  sources = [
    "source.amazon-ebs.ubuntu"
  ]

  provisioner "file" {
    source = "${var.deb}"
    destination = "${local.dest_deb}"
  }

  provisioner "shell" {
    inline = [
      "sudo dpkg -i ${local.dest_deb} 2>&1",
      "sudo apt-get update",
      "sudo apt-get install --yes sqlite3 jq  2>&1",
      "echo \"\nsource /opt/veraison/env/env.bash\" >> ~/.bashrc "
    ]
  }
}

# vim: set et sts=2 sw=2:
