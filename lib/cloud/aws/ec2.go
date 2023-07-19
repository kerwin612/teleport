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

package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/gravitational/teleport/api/types"
)

const (
	// awsInstanceStateName represents the state of the AWS EC2
	// instance - (pending | running | shutting-down | terminated | stopping | stopped )
	// https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instances.html
	// Used for filtering instances.
	awsInstanceStateName = "instance-state-name"

	// labelTeleportNodeName is the label key containing the Node name override.
	labelTeleportNodeName = types.TeleportNamespace + "/node_name"

	// tagKeyForName is the AWS Tag used to store the resource's name.
	tagKeyForName = "Name"
)

var (
	// FilterRunningEC2Instance is an EC2 DescribeInstances filter for running instances.
	FilterRunningEC2Instance = ec2Types.Filter{
		Name:   aws.String(awsInstanceStateName),
		Values: []string{string(ec2Types.InstanceStateNameRunning)},
	}
)

// EC2NameFromLabelWithOverride returns the EC2 Instance Name.
// If teleport.dev/node_name tag is present, it uses its value as Instance Name.
// If none of the tags is present, the fallbackName is returned
func EC2NameFromLabelWithOverride(labels map[string]string, fallbackName string) string {
	if name, override := labels[labelTeleportNodeName]; override {
		return name
	}
	if name, tagName := labels[tagKeyForName]; tagName {
		return name
	}

	return fallbackName
}
