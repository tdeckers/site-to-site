package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sns"
	"gopkg.in/ini.v1"
)

const (
	notificationSubject = "Site-to-site VPN"
)

// CfnEvent are events from SNS to this Lambda function
type CfnEvent struct {
	Records []struct {
		EventSource  string
		EventVersion string
		Sns          struct {
			MessageId string
			Message   string // Contains CfnMessage
		}
	}
}

// CfnMessage represents the CloudFormation event details passed over SNS.
//
type CfnMessage struct {
	LogicalResourceId    string
	ResourceType         string
	ResourceStatus       string
	ResourceStatusReason string
	ResourceProperties   string
	StackName            string
}

var sess = session.Must(session.NewSessionWithOptions(session.Options{
	SharedConfigState: session.SharedConfigEnable,
}))
var ec2Client = ec2.New(sess)
var snsClient = sns.New(sess)

/**
 * Lambda function handler.
 *
 */
func handler(event CfnEvent) error {
	if len(event.Records) > 1 {
		log.Printf("received %d records\n", len(event.Records))
	}
	if err := validateEnv(); err != nil {
		return err
	}

	// retrieve
	msg, err := ParseCfnMessage(event.Records[0].Sns.Message)
	if err != nil {
		return err
	}
	logf("%s (%s): %s\n", msg.ResourceType, msg.LogicalResourceId, msg.ResourceStatus)
	if msg.ResourceType == "AWS::CloudFormation::Stack" &&
		msg.StackName == os.Getenv("PREFIX") {
		if msg.ResourceStatus == "CREATE_IN_PROGRESS" {
			logf("Creating stack...")
			notify("Creating stack...")
		}
		// TODO: also handle "UPDATE_COMPLETE"
		if msg.ResourceStatus == "CREATE_COMPLETE" {
			// TODO: also handle "UPDATE_COMPLETE"
			logf("Stack created!")
			notify("Stack created, configuring Meraki.")
			// Get CIDR block from VPC
			vpcid := os.Getenv("VPC_ID")
			cidr, err := GetVpcCidr(vpcid)
			if err != nil {
				notify("Failed to get VPC CIDR")
				return err
			}
			// Get outside ip and secret from VPNConnection
			conn, err := GetVPNDetails(os.Getenv("PREFIX") + "-vpn")
			if err != nil {
				notify("Falied to get VPN details")
				return err
			}
			if len(conn.Tunnels) != 2 {
				notify("Failed to create VPN connection. Was expecting 2 tunnels.")
				return fmt.Errorf("expecting 2 tunnels, got %d", len(conn.Tunnels))
			}
			// Add/update peer in Meraki
			peer := NewPeer(conn.Tunnels[0].OutsideIPAddress, []string{cidr}, conn.Tunnels[0].SharedSecret)
			err = ConfigureMeraki(peer)
			if err != nil {
				notify(fmt.Sprintf("Failed to configure Meraki: %v", err))
			} else {
				notify("Meraki configured.")
			}

		}
		if msg.ResourceStatus == "DELETE_IN_PROGRESS" {
			logf("Deleting stack...")
		}
		if msg.ResourceStatus == "DELETE_COMPLETE" {
			logf("Stack deleted!")
			notify("Stack deleted!")
			// Remove VPN peer from Meraki
			peers, err := GetPeers()
			if err != nil {
				logf("Failed to get peers: %v", err)
				return err
			}
			peerName := fmt.Sprintf("%s-cloud-vpn", os.Getenv("PREFIX"))
			newPeers := make([]Peer, len(peers))
			for _, peer := range peers {
				if peer.Name != peerName {
					newPeers = append(newPeers, peer)
				}
			}
			err = UpdatePeers(newPeers)
			if err != nil {
				logf("Failed to update peers: %v", err)
				return err
			}
		}
		// TODO: handle CREATE_FAILED, DELETE_FAILED
	}
	return nil
}

func validateEnv() error {
	missing := []string{}
	if os.Getenv("PREFIX") == "" {
		missing = append(missing, "PREFIX")
	}
	if os.Getenv("VPC_ID") == "" {
		missing = append(missing, "VPC_ID")
	}
	if len(missing) != 0 {
		msg := strings.Join(missing, ", ")
		return fmt.Errorf("environment variables not set correctly. missing: %s", msg)
	}
	return nil
}

// ParseCfnMessage parses the CloudFormation ini-style data into a
// CfnMessage struct.
func ParseCfnMessage(m string) (*CfnMessage, error) {
	parsed, err := ini.Load([]byte(m))
	if err != nil {
		return &CfnMessage{}, err
	}
	msg := new(CfnMessage)
	err = parsed.MapTo(msg)
	if err != nil {
		return &CfnMessage{}, err
	}
	return msg, nil
}

// GetVpcCidr retrieves the CIDR block for a given AWS VPC.
func GetVpcCidr(vpcid string) (string, error) {
	input := &ec2.DescribeVpcsInput{VpcIds: []*string{&vpcid}}
	output, err := ec2Client.DescribeVpcs(input)
	if err != nil {
		return "", err
	}
	if len(output.Vpcs) > 1 { // Should never happen
		return "", errors.New("more than one VPC found with id")
	}
	return *output.Vpcs[0].CidrBlock, nil
}

// VPNConnection struct is used for parsing XML in the CustomerGatewayConfiguration
// of the VPNConnection retrieved from the AWS DescribeVpnConnection API
type VPNConnection struct {
	CustomerGatewayID string     `xml:"customer_gateway_id"`
	Tunnels           []struct { // typically 2 tunnels exist.
		OutsideIPAddress string `xml:"vpn_gateway>tunnel_outside_address>ip_address"`
		SharedSecret     string `xml:"ike>pre_shared_key"`
	} `xml:"ipsec_tunnel"`
}

// GetVPNDetails finds the VPNConnection with given name.
func GetVPNDetails(vpnConnectionName string) (*VPNConnection, error) {
	// search VPN based on the Name tag.
	input := &ec2.DescribeVpnConnectionsInput{
		Filters: []*ec2.Filter{&ec2.Filter{
			Name:   aws.String("tag:Name"),
			Values: []*string{aws.String(vpnConnectionName)},
		}},
	}
	output, err := ec2Client.DescribeVpnConnections(input)

	if err != nil {
		return nil, err
	}
	if len(output.VpnConnections) > 1 { // Should never happen
		return nil, errors.New("more than one VPNConnection found with id")
	}
	var conn = new(VPNConnection)
	if len(output.VpnConnections) < 1 {
		return nil, errors.New("no vpn connection details found")
	}
	xml.Unmarshal([]byte(*output.VpnConnections[0].CustomerGatewayConfiguration), &conn)
	for i, tunnel := range conn.Tunnels {
		logf("Tunnel %d: Outside IP: %s, Secret: %s\n", i, tunnel.OutsideIPAddress, tunnel.SharedSecret)
	}
	return conn, nil
}

// notify sends a message to an SNS topic, which is typically subscribed to
// by an admin/operator.
func notify(msg string) {
	topic := os.Getenv("NOTIFICATION_TOPIC")
	subject := notificationSubject
	input := &sns.PublishInput{
		Subject:  &subject,
		Message:  &msg,
		TopicArn: &topic,
	}
	_, err := snsClient.Publish(input)
	if err != nil {
		logf("Failed to send notification - error: %v\n", err)
	}
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
