terraform {
  backend "s3" {
    bucket = "site-to-site.app.ducbase.com"
    key    = "terraform.tfstate"
    region = "eu-west-1"
    #dynamodb_table = "terraform-lock"
  }
}
