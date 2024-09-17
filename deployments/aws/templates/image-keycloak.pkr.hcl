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

variable "keycloak_version" {
  type = string
  default = "25.0.5"
}

variable "conf_path" {
  type = string
}

variable "service_path" {
  type = string
}

locals {
    conf_dest = "/opt/keycloak/conf/keycloak.conf"
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
  name = "veraison-keycloak"
  sources = [
    "source.amazon-ebs.ubuntu"
  ]

  provisioner "file" {
    source = "${var.conf_path}"
    destination = "keycloak.conf"
  }

  provisioner "file" {
    source = "${var.service_path}"
    destination = "keycloak.service"
  }

  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get update", # doing it twice as once doesn't seem to be enough ....
      "sudo apt-get install -f --yes openjdk-21-jdk  2>&1",

      "sudo groupadd --system keycloak",
      "sudo useradd --system --gid keycloak --no-create-home --shell /bin/false keycloak",

      "wget https://github.com/keycloak/keycloak/releases/download/${var.keycloak_version}/keycloak-${var.keycloak_version}.tar.gz",
      "tar xf keycloak-${var.keycloak_version}.tar.gz",
      "rm keycloak-${var.keycloak_version}.tar.gz",
      "sudo mv keycloak-${var.keycloak_version} /opt/keycloak",
      "sudo mv keycloak.conf /opt/keycloak/conf/keycloak.conf",
      "sudo mv keycloak.service /opt/keycloak",
      "sudo mkdir -p /opt/keycloak/data/import",
      "sudo mkdir -p /opt/keycloak/certs",

      "sudo chown -R keycloak:keycloak /opt/keycloak",
      "sudo systemctl enable /opt/keycloak/keycloak.service",
    ]
  }
}

# vim: set et sts=2 sw=2:
