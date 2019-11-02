# Configure the AWS Provider
provider "aws" {
  #version = "~> 2.0"
  region  = "eu-west-1"
}

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}