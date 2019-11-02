package main

import (
	"fmt"
	"os"
	"testing"
)

func TestParseCfnMessage(t *testing.T) {
	debug = true
	rawMessage := `
	StackId='arn:aws:cloudformation:eu-west-1:014341863605:stack/s2s/20d4f1e0-f8e0-11e9-9b71-0a0cb138cf0a'
	Timestamp='2019-10-27T17:36:11.613Z'
	EventId='cgw0b574d54e07b6445d-CREATE_IN_PROGRESS-2019-10-27T17:36:11.613Z'
	LogicalResourceId='cgw0b574d54e07b6445d'
	Namespace='014341863605'
	ResourceProperties='{"Type": "ipsec.1",}'
	ResourceStatus='CREATE_IN_PROGRESS'
	ResourceStatusReason=''
	ResourceType='AWS::EC2::CustomerGateway'
	StackName='s2s'
	ClientRequestToken='null'`

	msg, err := ParseCfnMessage(rawMessage)
	if err != nil {
		t.Fatal(err)
	}
	if msg.ResourceStatus != "CREATE_IN_PROGRESS" {
		t.Fatal("failed to parse CfnMessage correctly")
	}
}

func TestGetVpc(t *testing.T) {
	debug = true

	out, err := GetVpcCidr("vpc-08b50691f17c334b5")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Result: %v\n", out)
}

func TestGetVpnDetails(t *testing.T) {
	debug = true

	out, err := GetVPNDetails("ducbase-vpn")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(out.Tunnels) != 2 {
		t.Errorf("Expecting 2 tunnels, found %d", len(out.Tunnels))
	}
}

func TestNotify(t *testing.T) {
	os.Setenv("NOTIFICATION_TOPIC", "arn:aws:sns:eu-west-1:014341863605:notify-me")
	notify("running unit test")
}
