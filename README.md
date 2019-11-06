# Site-to-site VPN Controller

## Overview

The code and configuration in this repo allows you to deploy a solution that can automatically set up and tear down a VPN connection between your home network and AWS.  With the VPN connection in place, you can directly access cloud resources from your home network over a secure connection.

The VPN Connection attaches to a specific VPC and route table in AWS.  Any resource (e.g. EC2 instance) provisioned in a subnet configured with the route table will be accessible.

This repo assumes Meraki MX to be the home side of the VPN Connection.  Other VPN solutions might be possible, but instructions are not included here.

## Prerequisites

* AWS account and CLI configured
* Terraform installed
* (optional) S3 bucket.  Local state can be used as well
* (optional) Meraki account and MX for home side of the VPN tunnel.  You can use alternative VPN solutions, but this repo doesn't provide instruction.

Before using terraform:
* Ensure AWS credentials are set.  Run `aws configure` if needed.
* Update `private.sh` with AWS and Meraki details
* verify and update `provider.tf` as needed.
* update `backend.tf` with your S3 bucket. If you want to use local state, you can remove the file.


## Deploy

1. Build lambda functions

    `make build`

2. (only needed once) Initialize terraform

    `make init`

3. Deploy infrastructure to AWS

    `source ./env.sh`
    `make deploy`

## Usage

TODO