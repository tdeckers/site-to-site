# Name tag applied to resources supporting tags
variable "tag_name" {
    default = "site-to-site"
}

# Prefix used for resource names that support a name or name tag
variable "prefix" {
    default = "s2s"
}

# VPC ID to attach the VPN connection to
variable "vpc_id" {
}

# ID of the route table to connect the VPN connection to
variable "route_table_id" {

}

# Public IP address of the Meraki MX gateway in the home network
variable "home_ip" {

}

# S3 bucket for terraform state and cloud formation template
variable "bucket_name" {
    
}

# SNS topic for sending progress notification to an admin/operator
# If empty, then no notifications will be sent.
variable "notification_topic" {
    default = ""
}

# Meraki org id for use in API
# Optional.  Set if you want to auto-update Meraki MX
variable "meraki_org_id" {
    default = ""
}

# Meraki API KEY
# Optional.  Set if you want to auto-update Meraki MX
variable "meraki_apikey" {
    default = ""
}