package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	merakiURL = "https://api.meraki.com/api/v0/organizations/%s/thirdPartyVPNPeers"
)

// Peer represents a 3rd party peer object to be used with the Meraki API
//
// [
//     {
//         "name": "aws-cisco",
//         "publicIp": "52.19.90.120",
//         "privateSubnets": [
//             "192.168.3.0/24"
//         ],
//         "secret": "yNCGShPMpLrKWxD7Moe2QR5pTPwAP3Ff",
//         "networkTags": [
//             "all"
//         ],
//         "ipsecPoliciesPreset": "aws"
//     }
// ]
type Peer struct {
	Name              string   `json:"name"`
	PublicIP          string   `json:"publicIp"`
	PrivateSubnets    []string `json:"privateSubnets"`
	Secret            string   `json:"secret"`
	NetworkTags       []string `json:"networkTags"`
	IpsecPolicyPreset string   `json:"ipsecPoliciesPreset"`
}

// NewPeer creates a new instance of Peer, with some defaults set for AWS vpns.
func NewPeer(publicIP string, subnets []string, secret string) *Peer {
	name := fmt.Sprintf("%s-cloud-vpn", os.Getenv("PREFIX"))
	return &Peer{
		Name:              name,
		PublicIP:          publicIP,
		PrivateSubnets:    subnets,
		Secret:            secret,
		NetworkTags:       []string{"all"},
		IpsecPolicyPreset: "aws",
	}
}

// GetPeers retrieve current 3rd party VPN peers using the Meraki API
func GetPeers() ([]Peer, error) {
	orgID := os.Getenv("MERAKI_ORG_ID")
	apiKey := os.Getenv("MERAKI_APIKEY")

	url := fmt.Sprintf(merakiURL, orgID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Cisco-Meraki-API-Key", apiKey)
	req.Header.Add("Accept", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	logf("BODY: %s\n", body)
	var peers []Peer
	if err := json.Unmarshal(body, &peers); err != nil {
		return nil, err
	}
	logf("# of peers: %d\n", len(peers))

	return peers, nil
}

// UpdatePeers updates the list of 3rd party VPN peers with the given peers list.
// This operation overwrites all existing peers. Make sure you first GetPeers and update
// from there.
func UpdatePeers(peers []Peer) error {
	orgID := os.Getenv("MERAKI_ORG_ID")
	apiKey := os.Getenv("MERAKI_APIKEY")

	url := fmt.Sprintf(merakiURL, orgID)

	jsonPeers, _ := json.Marshal(peers)
	logf("PUT body: %s", jsonPeers)
	req, _ := http.NewRequest("PUT", url, strings.NewReader(string(jsonPeers)))
	req.Header.Add("X-Cisco-Meraki-API-Key", apiKey)
	req.Header.Add("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		logf("HTTP %s: %s", res.Status, body)
		return fmt.Errorf("HTTP error code: %s", res.Status)
	}
	logf("Updated %d peers", len(peers))
	return nil
}

// ConfigureMeraki configures a 3rd party vpn peer for Meraki MX
//
// Expects MERAKI_ORGID and MERAKI_APIKEY to be set
func ConfigureMeraki(peer *Peer) error {
	orgID := os.Getenv("MERAKI_ORG_ID")
	apiKey := os.Getenv("MERAKI_APIKEY")
	if orgID == "" || apiKey == "" {
		return fmt.Errorf("MERAKI_ORG_ID and MERAKI_APIKEY must be set to configure Meraki MX")
	}

	peers, err := GetPeers()
	if err != nil {
		return err
	}

	// check if we configured the peer before. If `found` update, else create.
	found := false
	for i, p := range peers {
		if p.Name == peer.Name {
			peers[i] = *peer // replace existing peer
			found = true
		}
	}
	if !found {
		// add as new peer
		peers = append(peers, *peer)
	}

	if err := UpdatePeers(peers); err != nil {
		return err
	}

	return nil
}
