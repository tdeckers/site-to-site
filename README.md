[![published](https://static.production.devnetcloud.com/codeexchange/assets/images/devnet-published.svg)](https://developer.cisco.com/codeexchange/github/repo/tdeckers/site-to-site)
# Site-to-site VPN Controller

## Overview

The code and configuration in this repo allows you to deploy a solution that can automatically set up and tear down a VPN connection between your home network and AWS.  With the VPN connection in place, you can directly access cloud resources from your home network over a secure connection.

The VPN Connection attaches to a specific VPC and route table in AWS.  Any resource (e.g. EC2 instance) provisioned in a subnet configured with the route table will be accessible.

This repo assumes Meraki MX to be the home side of the VPN Connection.  Other VPN solutions might be possible, but instructions are not included here.

For a detailed overview of the purpose and use of this solution, check out [this post](https://medium.com/@ducbase/automate-your-home-network-extension-into-the-cloud-dbc6ed38bd4b).

## Prerequisites

* AWS account and CLI configured
* Terraform installed
* S3 bucket.  Used for remote state and cloud formation template.  For my setup requires public access of the S3 bucket.  You can change that.
* (optional) Meraki account and MX for home side of the VPN tunnel.  You can use alternative VPN solutions, but this repo doesn't provide instruction.

Before using terraform:

* Ensure AWS credentials are set.  Run `aws configure` if needed.
* Update `private.sh` with AWS and Meraki details
* verify and update `provider.tf` as needed.
* update `backend.tf` with your S3 bucket. If you want to use local state, you can remove the file.

## Deploy

1. Build lambda functions

```shell
    make build
```

2. (only needed once) Initialize terraform

```shell
    make init
```

3. Deploy infrastructure to AWS

```shell
    source ./env.sh
    make deploy
```

## Usage

The deployment will create a number of resources in your AWS account.  Most importantly, it'll create an API Gateway endpoint that you can trigger to create and delete the VPN connection.

Navigate to the API Gateway Console.  Under APIs, find *Site-to-site API*.  Click on *ANY* and then on *Test*.  This will open a page to trigger API calls.

To verify if a VPN connection is already create, select *Method* GET and click on *Test* at the bottom of the page.

To create a VPN connection, select *Method* POST and add `on` as the *Request Body*.  After about 10 minutes the VPN connection is created and optionally Meraki MX is configured.

To tear down the VPN connection, select *Method* POST and add `off` as the *Request Body*.
