package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

const (
	templateURL = "https://site-to-site.s3-eu-west-1.amazonaws.com/site-to-site.yaml"
)

var sess = session.Must(session.NewSessionWithOptions(session.Options{
	SharedConfigState: session.SharedConfigEnable,
}))
var cf = cloudformation.New(sess)

/**
 * Lambda function handler.
 * GET: retrieve state of the site-to-site VPN stack (Cloud formation): ON or OFF
 * POST:
 *  * body=ON  -- deploy site-to-site VPN stack (if doesn't exist already)
 *  * body=OFF -- destroy site-to-site VPN stack (if exist)
 */
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	method := request.HTTPMethod
	logf("method=%s\n", method)
	if err := validateEnv(); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	}

	stackName := os.Getenv("PREFIX")
	switch method {
	case "GET":
		stack, err := getStackByName(stackName)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				// No stack found
				return events.APIGatewayProxyResponse{
					Body:       "stack not found",
					StatusCode: 404,
				}, nil
			}
			return events.APIGatewayProxyResponse{}, err
		}
		return events.APIGatewayProxyResponse{
			Body:       *stack.StackStatus,
			StatusCode: 200,
		}, nil

	case "POST":
		logf("POST body: %s", request.Body)
		switch request.Body {
		case "ON":
			id, err := createStack(templateURL, stackName)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					return events.APIGatewayProxyResponse{
						Body:       err.Error(),
						StatusCode: 409,
					}, nil
				}
				return events.APIGatewayProxyResponse{}, err
			}
			return events.APIGatewayProxyResponse{
				Body:       id,
				StatusCode: 201,
			}, nil
		case "OFF":
			err := deleteStack(stackName)
			if err != nil {
				return events.APIGatewayProxyResponse{}, err
			}
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
			}, nil
		default:
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("post body not supported: %s", request.Body),
				StatusCode: 400,
			}, nil

		}
	default:
		return events.APIGatewayProxyResponse{
			Body:       "405 Method Not Allowed",
			StatusCode: 405, // method not allowed
		}, nil
	}
}

func validateEnv() error {
	missing := []string{}
	if os.Getenv("VPC_ID") == "" {
		missing = append(missing, "VPC_ID")
	}
	if os.Getenv("PREFIX") == "" {
		missing = append(missing, "PREFIX")
	}
	if os.Getenv("ROUTE_TABLE_ID") == "" {
		missing = append(missing, "ROUTE_TABLE_ID")
	}
	if os.Getenv("HOME_IP") == "" {
		missing = append(missing, "HOME_IP")
	}
	if len(missing) != 0 {
		msg := strings.Join(missing, ", ")
		return fmt.Errorf("environment variables not set correctly. missing: %s", msg)
	}
	return nil
}

func getStackByName(stackName string) (*cloudformation.Stack, error) {
	logf("Get stack with name: %s", stackName)
	input := &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)}
	output, err := cf.DescribeStacks(input)
	if err != nil {
		return nil, err
	}

	switch len(output.Stacks) {
	case 0: // should not happen
		return nil, fmt.Errorf("stack with name %s does not exist", stackName)
	case 1:
		stack := output.Stacks[0]
		logf("Status for stack %s: %s\n", stackName, *stack.StackStatus)
		return stack, nil
	default:
		return nil, fmt.Errorf("More than 1 stack found for %s", stackName)
	}
}

func createStack(url string, name string) (string, error) {
	vpcid := os.Getenv("VPC_ID")
	routeTableID := os.Getenv("ROUTE_TABLE_ID")
	homeIP := os.Getenv("HOME_IP")
	prefix := os.Getenv("PREFIX")

	params := []*cloudformation.Parameter{
		&cloudformation.Parameter{
			ParameterKey:   aws.String("VPCID"),
			ParameterValue: aws.String(vpcid),
		},
		&cloudformation.Parameter{
			ParameterKey:   aws.String("RouteTableID"),
			ParameterValue: aws.String(routeTableID),
		},
		&cloudformation.Parameter{
			ParameterKey:   aws.String("Prefix"),
			ParameterValue: aws.String(prefix),
		},
		&cloudformation.Parameter{
			ParameterKey:   aws.String("HomeIP"),
			ParameterValue: aws.String(homeIP),
		},
	}
	input := &cloudformation.CreateStackInput{
		TemplateURL: aws.String(url),
		StackName:   aws.String(name),
		Parameters:  params,
	}

	snsTopic := os.Getenv("SNS_TOPIC")
	if snsTopic != "" {
		input.SetNotificationARNs([]*string{&snsTopic})
	}

	output, err := cf.CreateStack(input)
	if err != nil {
		return "", err
	}

	return *output.StackId, nil
}

func deleteStack(name string) error {
	input := &cloudformation.DeleteStackInput{
		StackName: aws.String(name),
	}

	_, err := cf.DeleteStack(input)
	if err != nil {
		return err
	}
	return nil

}

var debug = true

func init() {
	debugEnv := os.Getenv("DEBUG")
	debug, _ = strconv.ParseBool(debugEnv)
}

func main() {
	lambda.Start(handler)
}

// helper function for logging
func logf(message string, v ...interface{}) {
	if debug {
		log.Printf(message, v...)
	}
}
