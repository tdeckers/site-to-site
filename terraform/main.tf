# Upload the CloudFormation template
resource "aws_s3_bucket_object" "s2s-vpn-template" {
  bucket = "${var.bucket_name}"
  key    = "site-to-site.yaml"
  source = "../cloudformation/site-to-site.yaml"
  acl    = "public-read"

  # The filemd5() function is available in Terraform 0.11.12 and later
  # For Terraform 0.11.11 and earlier, use the md5() function and the file() function:
  # etag = "${md5(file("path/to/file"))}"
  etag = "${filemd5("../cloudformation/site-to-site.yaml")}"
}

# Role for Lambda functions
resource "aws_iam_role" "role" {
  name = "s2s-lambda-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
  tags = {
    Application = "${var.tag_name}"
  }
}

# Define policy for all permissions required by both Lambda functions.
resource "aws_iam_policy" "policy" {
  name        = "s2s-lambda-policy"
  description = "Policy for site to site lambda functions"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "cloudformation:DescribeStacks",
        "cloudformation:CreateStack",
        "cloudformation:DeleteStack",
        "s3:GetObject",
        "SNS:Publish",
        "ec2:CreateVpnGateway",
        "ec2:DeleteVpnGateway",
        "ec2:DescribeVpnGateways",
        "ec2:AttachVpnGateway",
        "ec2:DetachVpnGateway",
        "ec2:EnableVgwRoutePropagation",
        "ec2:DisableVgwRoutePropagation",
        "ec2:DescribeCustomerGateways",
        "ec2:CreateCustomerGateway",
        "ec2:DeleteCustomerGateway",
        "ec2:createTags",
        "ec2:CreateVpnConnection",
        "ec2:DeleteVpnConnection",
        "ec2:DescribeVpnConnections",
        "ec2:DescribeRouteTables",
        "ec2:CreateVpnConnectionRoute",
        "ec2:DeleteVpnConnectionRoute",
        "ec2:DescribeVpcs"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

# Attach the policy to the role for Lambda functions
resource "aws_iam_role_policy_attachment" "role-policy-attach" {
  role       = "${aws_iam_role.role.name}"
  policy_arn = "${aws_iam_policy.policy.arn}"
}

# Function deploys/undeploys the site-to-site CloudFormation stack
resource "aws_lambda_function" "s2s-deploy" {
  filename      = "../bin/s2s-deploy.zip"
  function_name = "s2s-deploy"
  description = "Deploy/undeploy site-to-site CloudFormation stack"
  role          = "${aws_iam_role.role.arn}"
  handler       = "s2s-deploy"
  source_code_hash = "${filebase64sha256("../bin/s2s-deploy.zip")}"
  runtime = "go1.x"
  environment {
    variables = {
      DEBUG = "true"
      PREFIX = "${var.prefix}"
      VPC_ID = "${var.vpc_id}"
      ROUTE_TABLE_ID = "${var.route_table_id}"
      HOME_IP = "${var.home_ip}"
      BUCKET_NAME = "${var.bucket_name}"
      SNS_TOPIC = "${aws_sns_topic.s2s-event-topic.arn}"
    }
  }
  tags = {
    Application = "${var.tag_name}"
  }
}

# API gateway to handle HTTP requests to the deploy function
resource "aws_api_gateway_rest_api" "s2s-api" {
  name        = "Site-to-site API"
  description = "API for site-to-site VPN control"
}

# Reference to root resource of the API Gateway (root resource is /)
data "aws_api_gateway_resource" "s2s-resource" {
  rest_api_id = "${aws_api_gateway_rest_api.s2s-api.id}"
  path        = "/"
}

# Define the HTTP method to access the gateway.  Configured for ANY method.
# Note: deploy function only uses POST and GET
resource "aws_api_gateway_method" "s2s-method" {
  rest_api_id   = "${aws_api_gateway_rest_api.s2s-api.id}"
  resource_id   = "${aws_api_gateway_rest_api.s2s-api.root_resource_id}"
  http_method   = "ANY"
  authorization = "NONE"
}

# Integrate the API gateway method with the Lambda function
resource "aws_api_gateway_integration" "s2s-integration" {
  rest_api_id          = "${aws_api_gateway_rest_api.s2s-api.id}"
  resource_id          = "${aws_api_gateway_rest_api.s2s-api.root_resource_id}"
  http_method          = "${aws_api_gateway_method.s2s-method.http_method}"
  integration_http_method = "POST"
  type                 = "AWS_PROXY"
  uri                  = "${aws_lambda_function.s2s-deploy.invoke_arn}"
}

# Define permission for API Gateway to invoke the deploy function
resource "aws_lambda_permission" "s2s-apigw-lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.s2s-deploy.function_name}"
  principal     = "apigateway.amazonaws.com"

  # More: http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-control-access-using-iam-policies-to-invoke-api.html
  source_arn = "arn:aws:execute-api:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.s2s-api.id}/*"
}

# SNS topic that receives all CloudFormation site-to-site stack events.
# This is topic is configured for the stack by the deploy function.
resource "aws_sns_topic" "s2s-event-topic" {
  name = "s2s-stack-events"
}

# Subscribe event handler function to the SNS topic
resource "aws_sns_topic_subscription" "s2s-topic-subscription" {
  protocol = "lambda"
  topic_arn = "${aws_sns_topic.s2s-event-topic.arn}"
  endpoint  = "${aws_lambda_function.s2s-evt-handler.arn}"
}

# Function to handle events from the CloudFormation stack as it is deployed/undeployed.
resource "aws_lambda_function" "s2s-evt-handler" {
  filename      = "../bin/s2s-event-handler.zip"
  function_name = "s2s-event-handler"
  description = "Handle site-to-site stack create/delete events"
  role          = "${aws_iam_role.role.arn}"
  handler       = "s2s-event-handler"
  source_code_hash = "${filebase64sha256("../bin/s2s-event-handler.zip")}"
  runtime = "go1.x"
  environment {
    variables = {
      DEBUG = "true"
      PREFIX = "${var.prefix}"
      VPC_ID = "${var.vpc_id}"
      NOTIFICATION_TOPIC = "${var.notification_topic}"
      MERAKI_ORG_ID = "${var.meraki_org_id}"
      MERAKI_APIKEY = "${var.meraki_apikey}"
    }
  }
  tags = {
    Application = "${var.tag_name}"
  }
}

# Allow invocation of event-handler function from SNS.
resource "aws_lambda_permission" "s2s-sns-lambda" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.s2s-evt-handler.function_name}"
  principal     = "sns.amazonaws.com"
  source_arn    = "${aws_sns_topic.s2s-event-topic.arn}"
}

# Monitor tunnel uptime, notify when up for 12*600 secs (2 hours)
# Any value above 0 means something is up.. hence threshold = 0.2
resource "aws_cloudwatch_metric_alarm" "s2s-vpn-state" {
  alarm_name                = "s2s-vpn-tunnel-state"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "12"
  metric_name               = "TunnelState"
  namespace                 = "AWS/VPN"
  period                    = "600"
  statistic                 = "Maximum"
  threshold                 = "0.2"
  alarm_description         = "This metric monitors VPN tunnel state (up/down)"
  treat_missing_data        = "notBreaching"
  alarm_actions             = [length(var.notification_topic) > 0 ? var.notification_topic : ""]
  tags = {
    Application = "${var.tag_name}"
  }
}