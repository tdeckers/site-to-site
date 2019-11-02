# Site-to-site VPN Controller

## Prerequisites

* AWS CLI with credentials set
* Terraform
* S3 bucket for terraform state

## Deploy

Terraform resources are defined in the `terraform` directory.  State is kept in S3.

```
terraform init
```