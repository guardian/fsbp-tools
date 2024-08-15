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

func main() {

	ctx := context.Background()

	args := utils.ParseArgs()

	cfg, err := awsauth.LoadDefaultConfig(ctx, args.Profile, args.Region)
	if err != nil {
		log.Fatalf("%v", err)
	}

	securityHubClient := securityhub.NewFromConfig(cfg)

	ec2Client := ec2.NewFromConfig(cfg)

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

	groupDescriptions, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: securityGroups,
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
