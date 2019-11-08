#!/usr/bin/env bash

# This file contains configurations specific to your setup.
# It is added to .gitignore, so it can't be commited to git.

# Required
export VPC_ID=
export ROUTE_TABLE_ID=
export HOME_IP=
export BUCKET_NAME=

# Optional.  Set if you want Meraki MX to be auto-updated.
export MERAKI_ORG_ID=
export MERAKI_APIKEY=
# Optional. Set if you want notifications through SNS. (Set ARN here.)
export NOTIFICATION_TOPIC=
