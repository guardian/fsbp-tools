package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awsauth "github.com/guardian/fsbp-tools/aws-auth"
)

type cliArgs struct {
	Profile string
	Region  string
}

func ParseArgs() cliArgs {
	profile := flag.String("profile", "", "The name of the profile to use")
	region := flag.String("region", "", "The AWS region to use")

	flag.Parse()

	if *profile == "" {
		log.Fatal("Please provide a named AWS profile")
	}

	if *region == "" {
		log.Fatal("Please provide a region")
	}

	return cliArgs{
		Profile: *profile,
		Region:  *region,
	}
}

func idFromArn(arn string) string {
	splitArr := strings.Split(arn, "/")
	return splitArr[len(splitArr)-1]
}

func findTag(tags []types.Tag, key string, defaultValue string) string {
	for _, tag := range tags {
		if *tag.Key == key {
			return *tag.Value
		}
	}
	return defaultValue
}

func main() {

	ctx := context.Background()

	args := ParseArgs()

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
			sgId := idFromArn(*resource.Id)
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
			name := findTag(vpc.Tags, "Name", "unknown")
			fmt.Printf("Security group: %s, VPC name: %s, VPC id: %s\n", *group.GroupId, name, *group.VpcId)
		}
	}
}
