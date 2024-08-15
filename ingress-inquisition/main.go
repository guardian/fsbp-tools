package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awsauth "github.com/guardian/fsbp-tools/aws-common"
	"github.com/guardian/fsbp-tools/ingress-inquisition/utils"
)

func logGroupsAndVpcs(ctx context.Context, ec2Client *ec2.Client, groupIds []string) {
	groupDescriptions, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: groupIds,
	})
	if err != nil {
		log.Fatalf("Error describing security groups: %v", err)
	}

	for _, group := range groupDescriptions.SecurityGroups {
		vpcs, err := ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
			VpcIds: []string{*group.VpcId},
		})
		if err != nil {
			log.Fatalf("Error describing VPC: %v", err)
		}
		for _, vpc := range vpcs.Vpcs {
			name := utils.FindTag(vpc.Tags, "Name", "unknown")
			fmt.Printf("Security group: %s, VPC name: %s, VPC id: %s\n", *group.GroupId, name, *group.VpcId)
		}
	}
}

func main() {

	ctx := context.Background()

	args := utils.ParseArgs()

	cfg, err := awsauth.LoadDefaultConfig(ctx, args.Profile, args.Region)
	if err != nil {
		log.Fatalf("%v", err)
	}

	securityHubClient := securityhub.NewFromConfig(cfg)

	findings, err := awsauth.ReturnFindings(ctx, securityHubClient, "EC2.2", 100)
	if err != nil {
		log.Fatalf("Error getting findings: %v", err)
	}

	securityGroups := []string{}
	for _, finding := range findings.Findings {
		for _, resource := range finding.Resources {
			sgId := utils.IdFromArn(*resource.Id)
			securityGroups = append(securityGroups, sgId)
		}
	}

	ec2Client := ec2.NewFromConfig(cfg)

	logGroupsAndVpcs(ctx, ec2Client, securityGroups)

	maxInterfaceResults := int32(1000)

	securityGroupsInNetworkInterfaces := []string{}

	res, err := ec2Client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
		MaxResults: &maxInterfaceResults,
	})
	if err != nil {
		log.Fatalf("Error describing network interfaces: %v", err)
	}
	for _, networkInterface := range res.NetworkInterfaces {
		for _, group := range networkInterface.Groups {
			securityGroupsInNetworkInterfaces = append(securityGroupsInNetworkInterfaces, *group.GroupId)
		}
	}
}
