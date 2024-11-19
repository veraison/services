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

variable "command_dispatcher_path" {
  type = string
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
  name = "veraison-sentinel"
  sources = [
    "source.amazon-ebs.ubuntu"
  ]

  provisioner "file" {
    source = "${var.command_dispatcher_path}"
    destination = "veraison-dispatcher"
  }

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get update", # doing it twice as once doesn't seem to be enough ....
      "sudo NEEDRESTART_MODE=a apt-get install --yes postgresql gcc libpq-dev python3-venv python3-dev",
      "sudo systemctl disable postgresql",  # only need it to use psycopg2

      "sudo mkdir -p /opt/veraison/bin",
      "sudo mv veraison-dispatcher /opt/veraison/bin/veraison",
      "sudo chmod +x /opt/veraison/bin/veraison",
      "sudo ln -s /opt/veraison/bin/veraison /usr/bin/veraison",

      "sudo python3 -mvenv /opt/veraison/venv",
      "sudo /opt/veraison/venv/bin/pip install psycopg2 pyxdg",

      "sudo NEEDRESTART_MODE=a apt-get remove --yes gcc",
    ]
  }
}

# vim: set et sts=2 sw=2:
