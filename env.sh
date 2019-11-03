#!/usr/bin/env bash

# This file sources private.sh, which must not be commited to git since 
# it contains user specific details.
source ./private.sh

export TF_VAR_vpc_id="$VPC_ID"
export TF_VAR_route_table_id="$ROUTE_TABLE_ID"
export TF_VAR_home_ip="$HOME_IP"
export TF_VAR_meraki_org_id="$MERAKI_ORG_ID"
export TF_VAR_meraki_apikey="$MERAKI_APIKEY"
export TF_VAR_notification_topic="$NOTIFICATION_TOPIC"