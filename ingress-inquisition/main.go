package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/guardian/fsbp-tools/common"
	"github.com/guardian/fsbp-tools/ingress-inquisition/utils"
)

type SecurityGroupAndVpc struct {
	SecurityGroup string
	VpcName       string
	VpcId         string
}

func getVpcDetails(ctx context.Context, ec2Client *ec2.Client, groupIds []string) []SecurityGroupAndVpc {
	groupDescriptions, err := ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: groupIds,
	})
	if err != nil {
		log.Fatalf("Error describing security groups: %v", err)
	}

	res := []SecurityGroupAndVpc{}
	for _, group := range groupDescriptions.SecurityGroups {
		vpcs, err := ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
			VpcIds: []string{*group.VpcId},
		})
		if err != nil {
			log.Fatalf("Error describing VPC: %v", err)
		}
		for _, vpc := range vpcs.Vpcs {
			name := utils.FindTag(vpc.Tags, "Name", "unknown")
			res = append(res, SecurityGroupAndVpc{
				SecurityGroup: *group.GroupId,
				VpcName:       name,
				VpcId:         *group.VpcId,
			})
		}
	}
	return res
}

func findUnusedSecurityGroups(ctx context.Context, ec2Client *ec2.Client, sgIds []string) ([]string, error) {

	securityGroupsInNetworkInterfaces := []string{}

	maxInterfaceResults := int32(1000)
	res, err := ec2Client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
		MaxResults: &maxInterfaceResults,
	})
	if err != nil {
		return nil, err
	}
	for _, networkInterface := range res.NetworkInterfaces {
		for _, group := range networkInterface.Groups {
			securityGroupsInNetworkInterfaces = append(securityGroupsInNetworkInterfaces, *group.GroupId)
		}
	}

	return common.Complement(sgIds, securityGroupsInNetworkInterfaces), nil
}

func main() {

	ctx := context.Background()

	args := utils.ParseArgs()

	cfg, err := common.LoadDefaultConfig(ctx, args.Profile, args.Region)
	if err != nil {
		log.Fatalf("%v", err)
	}

	securityHubClient := securityhub.NewFromConfig(cfg)

	findings, err := common.ReturnFindings(ctx, securityHubClient, "EC2.2", 100)
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

	unusedSecurityGroups, err := findUnusedSecurityGroups(ctx, ec2Client, securityGroups)
	if err != nil {
		log.Fatalf("Error finding unused security groups: %v", err)
	}

	unusedSgVpcDetails := getVpcDetails(ctx, ec2Client, unusedSecurityGroups)

	fmt.Println("\nUnused default security groups with open ingress/egress:")

	// Print out results as a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Security Group\tVPC Name\tVPC ID")
	for _, sg := range unusedSgVpcDetails {
		fmt.Fprintf(w, "%s\t%s\t%s\n", sg.SecurityGroup, sg.VpcName, sg.VpcId)
	}
	w.Flush()
}
