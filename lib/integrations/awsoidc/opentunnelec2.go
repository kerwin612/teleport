/*
Copyright 2023 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package awsoidc

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	signer "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"github.com/gorilla/websocket"
	"github.com/gravitational/trace"
	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"

	"github.com/gravitational/teleport/lib/auth/native"
	"github.com/gravitational/teleport/lib/modules"
)

var (
	// validEC2Ports contains the available EC2 ports to use with EC2 Instance Connect Endpoint.
	validEC2Ports = []string{"22", "3389"}

	// filterEC2InstanceConnectEndpointStateKey is the filter key for filtering EC2 Instance Connection Endpoint by their state.
	filterEC2InstanceConnectEndpointStateKey = "state"
	// filterEC2InstanceConnectEndpointVPCIDKey is the filter key for filtering EC2 Instance Connection Endpoint by their VPC ID.
	filterEC2InstanceConnectEndpointVPCIDKey = "vpc-id"
)

// OpenTunnelEC2Request contains the required fields to open a tunnel to an EC2 instance.
// This will create a TCP socket that forwards incoming connections to the EC2's private IP address.
type OpenTunnelEC2Request struct {
	// Region is the AWS Region.
	Region string

	// InstanceID is the EC2 Instance's ID.
	InstanceID string

	// VPCID is the VPC where the EC2 Instance is located.
	// Used to look for the EC2 Instance Connect Endpoint.
	// Each VPC ID can only have one EC2 Instance Connect Endpoint.
	VPCID string

	// AvailabilityZoneID is the availability zone id where the instance is running.
	AvailabilityZoneID string

	// EC2SSHLoginUser is the OS user to use when the user wants SSH access.
	EC2SSHLoginUser string

	// EC2Address is the address to connect to in the EC2 Instance.
	// Eg, ip-172-31-32-234.eu-west-2.compute.internal:22
	EC2Address string

	Listener net.Listener

	// EC2Port is the port to connect to in the EC2 Instance.
	// This values is parsed from EC2Address.
	// Possible values: 22, 3389.
	ec2Port string
	// ec2PrivateIP is the private IP of the EC2 Instance.
	// This values is parsed from EC2Address.
	ec2PrivateIP string
}

// CheckAndSetDefaults checks if the required fields are present.
func (r *OpenTunnelEC2Request) CheckAndSetDefaults() error {
	var err error

	if r.Listener == nil {
		return trace.BadParameter("listener is required")
	}

	if r.InstanceID == "" {
		return trace.BadParameter("instance id is required")
	}

	if r.VPCID == "" {
		return trace.BadParameter("vpcid is required")
	}

	if r.EC2SSHLoginUser == "" {
		return trace.BadParameter("ec2 ssh login user is required")
	}

	if r.AvailabilityZoneID == "" {
		return trace.BadParameter("availability zone id is required")
	}

	if r.EC2Address == "" {
		return trace.BadParameter("ec2 address required")
	}

	r.ec2PrivateIP, r.ec2Port, err = net.SplitHostPort(r.EC2Address)
	if err != nil {
		return trace.BadParameter("ec2 address is invalid: %v", err)
	}
	if !slices.Contains(validEC2Ports, r.ec2Port) {
		return trace.BadParameter("invalid ec2 address port %s, possible values: %v", r.ec2Port, validEC2Ports)
	}

	return nil
}

// OpenTunnelEC2Response contains the response for creating a Tunnel to an EC2 Instance.
// It returns the listening address and the SSH Private Key (PEM encoded).
type OpenTunnelEC2Response struct {
	// SSHSigner is the SSH Signer that should be used to connect to the host.
	SSHSigner ssh.Signer
}

// OpenTunnelEC2Client describes the required methods to Open a Tunnel to an EC2 Instance using
// EC2 Instance Connect Endpoint.
type OpenTunnelEC2Client interface {
	// Describes the specified EC2 Instance Connect Endpoints or all EC2 Instance
	// Connect Endpoints.
	DescribeInstanceConnectEndpoints(ctx context.Context, params *ec2.DescribeInstanceConnectEndpointsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstanceConnectEndpointsOutput, error)

	// SendSSHPublicKey pushes an SSH public key to the specified EC2 instance for use by the specified
	// user. The key remains for 60 seconds. For more information, see Connect to your
	// Linux instance using EC2 Instance Connect (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Connect-using-EC2-Instance-Connect.html)
	// in the Amazon EC2 User Guide.
	SendSSHPublicKey(ctx context.Context, params *ec2instanceconnect.SendSSHPublicKeyInput, optFns ...func(*ec2instanceconnect.Options)) (*ec2instanceconnect.SendSSHPublicKeyOutput, error)

	// Retrieve returns nil if it successfully retrieved the value.
	// Error is returned if the value were not obtainable, or empty.
	Retrieve(ctx context.Context) (aws.Credentials, error)
}

type defaultOpenTunnelEC2Client struct {
	*ec2.Client
	awsCredentialsProvider aws.CredentialsProvider
	ec2icClient            *ec2instanceconnect.Client
}

// SendSSHPublicKey pushes an SSH public key to the specified EC2 instance for use by the specified
// user. The key remains for 60 seconds. For more information, see Connect to your
// Linux instance using EC2 Instance Connect (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Connect-using-EC2-Instance-Connect.html)
// in the Amazon EC2 User Guide.
func (d defaultOpenTunnelEC2Client) SendSSHPublicKey(ctx context.Context, params *ec2instanceconnect.SendSSHPublicKeyInput, optFns ...func(*ec2instanceconnect.Options)) (*ec2instanceconnect.SendSSHPublicKeyOutput, error) {
	return d.ec2icClient.SendSSHPublicKey(ctx, params, optFns...)
}

// Retrieve returns nil if it successfully retrieved the value.
// Error is returned if the value were not obtainable, or empty.
func (d defaultOpenTunnelEC2Client) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return d.awsCredentialsProvider.Retrieve(ctx)
}

// NewOpenTunnelEC2Client creates a OpenTunnelEC2Client using AWSClientRequest.
func NewOpenTunnelEC2Client(ctx context.Context, clientReq *AWSClientRequest) (OpenTunnelEC2Client, error) {
	ec2Client, err := newEC2Client(ctx, clientReq)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	ec2instanceconnectClient, err := newEC2InstanceConnectClient(ctx, clientReq)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	awsCredProvider, err := newAWSCredentialsProvider(ctx, clientReq)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &defaultOpenTunnelEC2Client{
		Client:                 ec2Client,
		awsCredentialsProvider: awsCredProvider,
		ec2icClient:            ec2instanceconnectClient,
	}, nil
}

// OpenTunnelEC2 creates a tunnel to an ec2 instance using its private IP.
// Ref:
// - https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/connect-using-eice.html
// - https://github.com/aws/aws-cli/blob/f6c820e89d8b566ab54ab9d863754ec4b713fd6a/awscli/customizations/ec2instanceconnect/opentunnel.py
func OpenTunnelEC2(ctx context.Context, clt OpenTunnelEC2Client, req OpenTunnelEC2Request) (*OpenTunnelEC2Response, error) {
	var err error
	ret := &OpenTunnelEC2Response{}

	if err := req.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	ret.SSHSigner, err = sendSSHPublicKey(ctx, clt, req)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// open websocket and let app connect to it
	dialer, err := NewDialer(ctx, clt, req)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// listener, err := net.Listen("tcp", ":0")
	// if err != nil {
	// 	return nil, trace.Wrap(err)
	// }
	listener := req.Listener

	log.Printf("Tunnel running on %s ...", listener.Addr().String())

	go func() {
		for {
			local, err := listener.Accept()
			if err != nil {
				log.Printf("failed to accept new ec2 iec tunnel conection: %v", err)
				return
			}

			log.Printf("got a new client: local=%v remote=%v \n", local.LocalAddr(), local.RemoteAddr())

			dialAddress := req.EC2Address
			remote, err := dialer.DialContext(ctx, "tcp", dialAddress)
			if err != nil {
				log.Printf("failed to dial remote on port 22: %v", err)
			}

			go func() {
				g, _ := errgroup.WithContext(ctx)

				g.Go(func() error {
					_, err := io.Copy(local, remote)
					if err != nil {
						return trace.Wrap(err, "remote->local")
					}

					return nil
				})

				g.Go(func() error {
					_, err := io.Copy(remote, local)
					if err != nil {
						return trace.Wrap(err, "local->remote")
					}

					return nil
				})

				err := g.Wait()
				if err != nil {
					log.Printf("aws oidc opentunnel - failed to copy data: instance=%s: %v", req.InstanceID, err)
				}
			}()
		}
	}()

	return ret, nil
}

func generatePrivatePublicKey() (publicKey any, privateKey any, err error) {
	if modules.GetModules().IsBoringBinary() {
		privKey, err := native.GeneratePrivateKey()
		if err != nil {
			return nil, nil, trace.Wrap(err)
		}

		return privKey.Public(), privKey, nil
	}

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	return pubKey, privKey, nil
}

// sendSSHPublicKey creates a new Private Key and uploads the Public to the ec2 instance.
// This key is only valid for 60 seconds and can only be used to authenticate the EC2SSHLoginUser.
func sendSSHPublicKey(ctx context.Context, clt OpenTunnelEC2Client, req OpenTunnelEC2Request) (ssh.Signer, error) {
	pubKey, privKey, err := generatePrivatePublicKey()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	publicKey, err := ssh.NewPublicKey(pubKey)

	pubKeySSH := string(ssh.MarshalAuthorizedKey(publicKey))
	_, err = clt.SendSSHPublicKey(ctx,
		&ec2instanceconnect.SendSSHPublicKeyInput{
			InstanceId:       &req.InstanceID,
			InstanceOSUser:   &req.EC2SSHLoginUser,
			SSHPublicKey:     &pubKeySSH,
			AvailabilityZone: &req.AvailabilityZoneID,
		},
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	sshSigner, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return sshSigner, nil
}

// sendSSHPublicKey creates a new Private Key and uploads the Public to the ec2 instance.
// This key is only valid for 60 seconds and can only be used to authenticate the EC2SSHLoginUser.
func sendSSHPublicKeyOld(ctx context.Context, clt OpenTunnelEC2Client, req OpenTunnelEC2Request) (ssh.Signer, error) {
	rsaPrivateKey, err := native.GeneratePrivateKey()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	pubKeySSH := string(rsaPrivateKey.MarshalSSHPublicKey())
	_, err = clt.SendSSHPublicKey(ctx,
		&ec2instanceconnect.SendSSHPublicKeyInput{
			InstanceId:       &req.InstanceID,
			InstanceOSUser:   &req.EC2SSHLoginUser,
			SSHPublicKey:     &pubKeySSH,
			AvailabilityZone: &req.AvailabilityZoneID,
		},
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	sshSigner, err := ssh.NewSignerFromKey(rsaPrivateKey)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return sshSigner, nil
}

type Dialer struct {
	clt              aws.CredentialsProvider
	awsRegion        string
	endpointId       string
	endpointDns      string
	privateIPAddress string
	remotePort       string
}

func NewDialer(ctx context.Context, clt OpenTunnelEC2Client, req OpenTunnelEC2Request) (*Dialer, error) {
	describe, err := clt.DescribeInstanceConnectEndpoints(ctx, &ec2.DescribeInstanceConnectEndpointsInput{
		Filters: []ec2types.Filter{
			{
				Name:   &filterEC2InstanceConnectEndpointVPCIDKey,
				Values: []string{req.VPCID},
			},
			{
				Name:   &filterEC2InstanceConnectEndpointStateKey,
				Values: []string{string(ec2types.Ec2InstanceConnectEndpointStateCreateComplete)},
			},
		},
	})
	if err != nil {
		return nil, trace.BadParameter("describing ec2 instance connect endpoint: %v", err)
	}

	if len(describe.InstanceConnectEndpoints) == 0 {
		return nil, trace.BadParameter("no instance connect endpoint is available")
	}

	// Use the first that is returned.
	ec2ice := describe.InstanceConnectEndpoints[0]

	return &Dialer{
		clt:              clt,
		awsRegion:        req.Region,
		endpointId:       *ec2ice.InstanceConnectEndpointId,
		endpointDns:      *ec2ice.DnsName,
		privateIPAddress: req.ec2PrivateIP,
		remotePort:       req.ec2Port,
	}, nil
}

func (icd *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/connect-using-eice.html
	maxTunnelDurationInSeconds := "3600" // 1 hour

	if network != "tcp" && network != "tcp4" {
		return nil, trace.BadParameter("only tcp supported")
	}

	q := url.Values{}
	q.Set("instanceConnectEndpointId", icd.endpointId)
	q.Set("maxTunnelDuration", maxTunnelDurationInSeconds)
	q.Set("privateIpAddress", icd.privateIPAddress)
	q.Set("remotePort", icd.remotePort)

	openTunnelURL := url.URL{
		Scheme:   "wss",
		Host:     icd.endpointDns,
		Path:     "openTunnel",
		RawQuery: q.Encode(),
	}

	r, _ := http.NewRequest(http.MethodGet, openTunnelURL.String(), nil)

	sum := sha256.Sum256([]byte{})
	sumStr := hex.EncodeToString(sum[:])

	creds, err := icd.clt.Retrieve(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	s := signer.NewSigner()
	signed, _, err := s.PresignHTTP(ctx, creds, r, sumStr, "ec2-instance-connect", icd.awsRegion, time.Now())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, signed, http.Header{})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &icdConn{
		Conn: conn,
		r:    websocket.JoinMessages(conn, ""),
	}, nil
}

type icdConn struct {
	*websocket.Conn
	r io.Reader
}

var _ net.Conn = (*icdConn)(nil)

func (i *icdConn) Read(b []byte) (n int, err error) {
	return i.r.Read(b)
}

func (i *icdConn) Write(b []byte) (n int, err error) {
	err = i.Conn.WriteMessage(websocket.BinaryMessage, b)
	n = len(b)
	return
}

func (i *icdConn) SetDeadline(t time.Time) error {
	i.SetReadDeadline(t)
	i.SetWriteDeadline(t)
	return nil
}
